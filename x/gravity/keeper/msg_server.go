package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/errors"
	sdkerrors "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	utilkeeper "github.com/palomachain/paloma/util/keeper"
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

// SendToEth handles MsgSendToEth
func (k msgServer) SendToEth(c context.Context, msg *types.MsgSendToEth) (*types.MsgSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid sender")
	}
	dest, err := types.NewEthAddress(msg.EthDest)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid eth dest")
	}
	erc20, err := k.GetERC20OfDenom(ctx, msg.GetChainReferenceId(), msg.Amount.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid denom")
	}
	if k.InvalidSendToEthAddress(ctx, *dest, *erc20) {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "destination address is invalid")
	}
	txID, err := k.AddToOutgoingPool(ctx, sender, *dest, msg.Amount, msg.GetChainReferenceId())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not add to outgoing pool")
	}
	return &types.MsgSendToEthResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventOutgoingTxId{
			Message: msg.Type(),
			TxId:    fmt.Sprint(txID),
		},
	)
}

// ConfirmBatch handles MsgConfirmBatch
func (k msgServer) ConfirmBatch(c context.Context, msg *types.MsgConfirmBatch) (*types.MsgConfirmBatchResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid MsgConfirmBatch")
	}
	contract, err := types.NewEthAddress(msg.TokenContract)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "eth address invalid")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// fetch the outgoing batch given the nonce
	batch, err := k.GetOutgoingTXBatch(ctx, *contract, msg.Nonce)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find batch")
	}

	ci, err := k.evmKeeper.GetChainInfo(ctx, batch.ChainReferenceID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find chain info")
	}
	checkpoint, err := batch.GetCheckpoint(string(ci.SmartContractUniqueID))
	if err != nil {
		return nil, err
	}

	orchaddr, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}

	err = k.confirmHandlerCommon(ctx, msg.EthSigner, orchaddr, msg.Signature, checkpoint, batch.ChainReferenceID)
	if err != nil {
		return nil, err
	}

	// check if we already have this confirm
	batchConfirms, err := k.GetBatchConfirm(ctx, msg.Nonce, *contract, orchaddr)
	if err != nil {
		return nil, err
	}
	if batchConfirms != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "duplicate signature")
	}
	key, err := k.SetBatchConfirm(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgConfirmBatchResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventBatchConfirmKey{
			Message:         msg.Type(),
			BatchConfirmKey: string(key),
		},
	)
}

// checkOrchestratorValidatorInSet checks that the orchestrator refers to a validator that is
// currently in the set
func (k msgServer) checkOrchestratorValidatorInSet(ctx context.Context, orchestrator string) error {
	orchaddr, err := sdk.AccAddressFromBech32(orchestrator)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}
	validator, found, err := k.GetOrchestratorValidator(ctx, orchaddr)
	if err != nil {
		return err
	}
	if !found {
		return sdkerrors.Wrap(types.ErrUnknown, "validator")
	}
	valAddress, err := utilkeeper.ValAddressFromBech32(k.AddressCodec, validator.GetOperator())
	if err != nil {
		return err
	}
	// return an error if the validator isn't in the active set
	val, err := k.StakingKeeper.Validator(ctx, valAddress)
	if err != nil {
		return err
	}
	if val == nil || !val.IsBonded() {
		return sdkerrors.Wrap(errorsmod.ErrorInvalidSigner, "validator not in active set")
	}

	return nil
}

// claimHandlerCommon is an internal function that provides common code for processing claims once they are
// translated from the message to the Ethereum claim interface
func (k msgServer) claimHandlerCommon(ctx context.Context, msgAny *codectypes.Any, msg types.EthereumClaim) error {
	// Add the claim to the store
	_, err := k.Attest(ctx, msg, msgAny)
	if err != nil {
		return sdkerrors.Wrap(err, "create attestation")
	}
	hash, err := msg.ClaimHash()
	if err != nil {
		return sdkerrors.Wrap(err, "unable to compute claim hash")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Emit the handle message event
	return sdkCtx.EventManager().EmitTypedEvent(
		&types.EventClaim{
			Message:       msg.GetType().String(),
			ClaimHash:     fmt.Sprintf("%x", hash),
			AttestationId: fmt.Sprintf("%x", types.GetAttestationKey(msg.GetEventNonce(), hash)),
		},
	)
}

// confirmHandlerCommon is an internal function that provides common code for processing claim messages
func (k msgServer) confirmHandlerCommon(ctx context.Context, ethAddress string, orchestrator sdk.AccAddress, signature string, checkpoint []byte, chainReferenceId string) error {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "signature decoding")
	}

	submittedEthAddress, err := types.NewEthAddress(ethAddress)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "invalid eth address")
	}

	validator, found, err := k.GetOrchestratorValidator(ctx, orchestrator)
	if err != nil {
		return err
	}
	if !found {
		return sdkerrors.Wrap(types.ErrUnknown, "validator")
	}
	valAddress, err := utilkeeper.ValAddressFromBech32(k.AddressCodec, validator.GetOperator())
	if !validator.IsBonded() && !validator.IsUnbonding() {
		// We must only accept confirms from bonded or unbonding validators
		return sdkerrors.Wrap(types.ErrInvalid, "validator is unbonded")
	}

	if err := sdk.VerifyAddressFormat(valAddress); err != nil {
		return sdkerrors.Wrapf(err, "discovered invalid validator address for orchestrator %v", orchestrator)
	}

	ethAddressFromStore, found, err := k.GetEthAddressByValidator(ctx, valAddress, chainReferenceId)
	if err != nil {
		return err
	}
	if !found {
		return sdkerrors.Wrap(types.ErrEmpty, "no eth address set for validator")
	}

	if *ethAddressFromStore != *submittedEthAddress {
		return sdkerrors.Wrap(types.ErrInvalid, "submitted eth address does not match delegate eth address")
	}

	err = types.ValidateEthereumSignature(checkpoint, sigBytes, *ethAddressFromStore)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("signature verification failed expected sig by %s with checkpoint %s found %s", ethAddress, hex.EncodeToString(checkpoint), signature))
	}

	return nil
}

func (k msgServer) SendToPalomaClaim(c context.Context, msg *types.MsgSendToPalomaClaim) (*types.MsgSendToPalomaClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchstrator validator inset")
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check Any value")
	}
	err = k.claimHandlerCommon(ctx, any, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendToPalomaClaimResponse{}, nil
}

func (k msgServer) BatchSendToEthClaim(c context.Context, msg *types.MsgBatchSendToEthClaim) (*types.MsgBatchSendToEthClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchestrator validator")
	}

	/* Perform some additional checks on the input to determine if it is valid before allowing it on the chain
	   Note that because of the gas meter we must avoid calls which consume gas, like fetching data from the keeper
	*/
	err = additionalPatchChecks(ctx, k, msg)
	if err != nil {
		return nil, err
	}

	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check Any value")
	}

	err = k.claimHandlerCommon(ctx, msgAny, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgBatchSendToEthClaimResponse{}, nil
}

// Performs additional checks on msg to determine if it is valid
func additionalPatchChecks(ctx context.Context, k msgServer, msg *types.MsgBatchSendToEthClaim) error {
	contractAddress, err := types.NewEthAddress(msg.TokenContract)
	if err != nil {
		return sdkerrors.Wrap(err, "Invalid TokenContract on MsgBatchSendToEthClaim")
	}

	// Replicate the following but without using a gas meter
	b, err := k.GetOutgoingTXBatch(ctx, *contractAddress, msg.BatchNonce)
	if err != nil {
		return err
	}
	if b == nil {
		// Batch deleted, just add the vote to the stored attestation
		return nil
	}

	if b.BatchTimeout <= msg.EthBlockHeight {
		return fmt.Errorf("batch with nonce %d submitted after it timed out (submission %d >= timeout %d)", msg.BatchNonce, msg.EthBlockHeight, b.BatchTimeout)
	}

	return nil
}

func (k msgServer) CancelSendToEth(c context.Context, msg *types.MsgCancelSendToEth) (*types.MsgCancelSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}
	err = k.RemoveFromOutgoingPoolAndRefund(ctx, msg.TransactionId, sender)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelSendToEthResponse{}, nil
}

func (k msgServer) SubmitBadSignatureEvidence(c context.Context, msg *types.MsgSubmitBadSignatureEvidence) (*types.MsgSubmitBadSignatureEvidenceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.CheckBadSignatureEvidence(ctx, msg, msg.GetChainReferenceId())
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitBadSignatureEvidenceResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventBadSignatureEvidence{
			Message:                fmt.Sprint(msg.Type()),
			BadEthSignature:        fmt.Sprint(msg.Signature),
			BadEthSignatureSubject: fmt.Sprint(msg.Subject),
		},
	)
}

func (k msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
