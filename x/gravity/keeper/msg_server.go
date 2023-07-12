package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/palomachain/paloma/x/gravity/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) SetDelegateKeys(c context.Context, msg *types.MsgDelegateKeys) (*types.MsgDelegateKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	orchAddr, err := sdk.AccAddressFromBech32(msg.OrchestratorAddress)
	if err != nil {
		return nil, err
	}

	ethAddr := common.HexToAddress(msg.EthereumAddress)

	// ensure that the validator exists
	if k.Keeper.StakingKeeper.Validator(ctx, valAddr) == nil {
		return nil, sdkerrors.Wrap(stakingtypes.ErrNoValidatorFound, valAddr.String())
	}

	// check if the Ethereum address is currently not used
	if k.validatorForEthAddressExists(ctx, ethAddr) {
		return nil, sdkerrors.Wrapf(types.ErrDelegateKeys, "ethereum address %s in use", ethAddr)
	}

	// check if the orchestrator address is currently not used
	if k.ethAddressForOrchestratorExists(ctx, orchAddr) {
		return nil, sdkerrors.Wrapf(types.ErrDelegateKeys, "orchestrator address %s in use", orchAddr)
	}

	valAccAddr := sdk.AccAddress(valAddr)
	valAccSeq, err := k.accountKeeper.GetSequence(ctx, valAccAddr)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrDelegateKeys, "failed to get sequence for validator account %s", valAccAddr)
	}

	var nonce uint64
	if valAccSeq > 0 {
		nonce = valAccSeq - 1
	}

	signMsgBz := k.cdc.MustMarshal(&types.DelegateKeysSignMsg{
		ValidatorAddress: valAddr.String(),
		// We decrement since we process the message after the ante-handler which
		// increments the nonce.
		Nonce: nonce,
	})

	hash := crypto.Keccak256Hash(signMsgBz).Bytes()

	if err = types.ValidateEthereumSignature(hash, msg.EthSignature, ethAddr); err != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrDelegateKeys,
			"failed to validate delegate keys signature for Ethereum address %X; %s ;%d",
			ethAddr, err, nonce,
		)
	}

	k.SetOrchestratorValidatorAddress(ctx, valAddr, orchAddr)
	k.setValidatorEthereumAddress(ctx, valAddr, ethAddr)
	k.setEthereumOrchestratorAddress(ctx, ethAddr, orchAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeySetOrchestratorAddr, orchAddr.String()),
			sdk.NewAttribute(types.AttributeKeySetEthereumAddr, ethAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyValidatorAddr, valAddr.String()),
		),
	)

	return &types.MsgDelegateKeysResponse{}, nil

}

// SubmitEthereumTxConfirmation handles MsgSubmitEthereumTxConfirmation
func (k msgServer) SubmitEthereumTxConfirmation(c context.Context, msg *types.MsgSubmitEthereumTxConfirmation) (*types.MsgSubmitEthereumTxConfirmationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	confirmation, err := types.UnpackConfirmation(msg.Confirmation)
	if err != nil {
		return nil, err
	}

	val, err := k.getSignerValidator(ctx, msg.Signer)
	if err != nil {
		return nil, err
	}

	otx := k.GetOutgoingTx(ctx, confirmation.GetStoreIndex())
	if otx == nil {
		k.Logger(ctx).Error(
			"no outgoing tx",
			"store index", fmt.Sprintf("%x", confirmation.GetStoreIndex()),
		)
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find outgoing tx")
	}

	gravityID := k.getGravityID(ctx)
	checkpoint := otx.GetCheckpoint([]byte(gravityID))

	ethAddress := k.GetValidatorEthereumAddress(ctx, val)
	if ethAddress != confirmation.GetSigner() {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "eth address does not match signer eth address")
	}

	if err = types.ValidateEthereumSignature(checkpoint, confirmation.GetSignature(), ethAddress); err != nil {
		k.Logger(ctx).Error("error validating signature",
			"eth addr", ethAddress.String(),
			"gravityID", gravityID,
			"checkpoint", hex.EncodeToString(checkpoint),
			"type url", msg.Confirmation.TypeUrl,
			"signature", hex.EncodeToString(confirmation.GetSignature()),
			"error", err)
		return nil, sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf(
			"signature verification failed ethAddress %s gravityID %s checkpoint %s typeURL %s signature %s err %s",
			ethAddress.Hex(),
			gravityID,
			hex.EncodeToString(checkpoint),
			msg.Confirmation.TypeUrl,
			hex.EncodeToString(confirmation.GetSignature()),
			err,
		))
	}
	// TODO: should validators be able to overwrite their signatures?
	if k.getEthereumSignature(ctx, confirmation.GetStoreIndex(), val) != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "signature duplicate")
	}

	key := k.SetEthereumSignature(ctx, confirmation, val)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeyEthereumSignatureKey, string(key)),
		),
	)
	return &types.MsgSubmitEthereumTxConfirmationResponse{}, nil
}

// SubmitEthereumEvent handles MsgSubmitEthereumEvent
func (k msgServer) SubmitEthereumEvent(c context.Context, msg *types.MsgSubmitEthereumEvent) (*types.MsgSubmitEthereumEventResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	event, err := types.UnpackEvent(msg.Event)
	if err != nil {
		return nil, err
	}

	// return an error if the validator isn't in the active set
	val, err := k.getSignerValidator(ctx, msg.Signer)
	if err != nil {
		return nil, err
	}

	// Add the claim to the store
	_, err = k.recordEventVote(ctx, event, val)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create event vote record")
	}

	// Emit the handle message event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, fmt.Sprintf("%T", event)),
			// TODO: maybe return something better here? is this the right string representation?
			sdk.NewAttribute(types.AttributeKeyEthereumEventVoteRecordID, string(types.MakeEthereumEventVoteRecordKey(event.GetEventNonce(), event.Hash()))),
		),
	)

	return &types.MsgSubmitEthereumEventResponse{}, nil
}

// SendToEthereum handles MsgSendToEthereum
func (k msgServer) SendToEthereum(c context.Context, msg *types.MsgSendToEthereum) (*types.MsgSendToEthereumResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// ensure the denoms provided in the message will map correctly if they are gravity denoms
	types.NormalizeCoinDenom(&msg.Amount)
	types.NormalizeCoinDenom(&msg.BridgeFee)

	txID, err := k.createSendToEthereum(ctx, sender, msg.EthereumRecipient, msg.Amount, msg.BridgeFee)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents([]sdk.Event{
		sdk.NewEvent(
			types.EventTypeBridgeWithdrawalReceived,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
			sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
			sdk.NewAttribute(types.AttributeKeyOutgoingTXID, strconv.Itoa(int(txID))),
			sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(txID)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeyOutgoingTXID, fmt.Sprint(txID)),
		),
	})

	return &types.MsgSendToEthereumResponse{Id: txID}, nil
}

func (k msgServer) CancelSendToEthereum(c context.Context, msg *types.MsgCancelSendToEthereum) (*types.MsgCancelSendToEthereumResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.Keeper.cancelSendToEthereum(ctx, msg.Id, msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents([]sdk.Event{
		sdk.NewEvent(
			types.EventTypeBridgeWithdrawCanceled,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
			sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeyOutgoingTXID, fmt.Sprint(msg.Id)),
		),
	})

	return &types.MsgCancelSendToEthereumResponse{}, nil
}

func (k msgServer) SubmitEthereumHeightVote(c context.Context, msg *types.MsgEthereumHeightVote) (*types.MsgEthereumHeightVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	val, err := k.getSignerValidator(ctx, msg.Signer)
	if err != nil {
		return nil, err
	}

	k.Keeper.SetEthereumHeightVote(ctx, val, msg.EthereumHeight)

	return &types.MsgEthereumHeightVoteResponse{}, nil
}

// getSignerValidator takes an sdk.AccAddress that represents either a validator or orchestrator address and returns
// the assoicated validator address
func (k Keeper) getSignerValidator(ctx sdk.Context, signerString string) (sdk.ValAddress, error) {
	signer, err := sdk.AccAddressFromBech32(signerString)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "signer address")
	}
	var validatorI stakingtypes.ValidatorI
	if validator := k.GetOrchestratorValidatorAddress(ctx, signer); validator == nil {
		validatorI = k.StakingKeeper.Validator(ctx, sdk.ValAddress(signer))
	} else {
		validatorI = k.StakingKeeper.Validator(ctx, validator)
	}

	if validatorI == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "not orchestrator or validator")
	} else if !validatorI.IsBonded() {
		return nil, sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("validator is not bonded: %s", validatorI.GetOperator()))
	}

	return validatorI.GetOperator(), nil
}
