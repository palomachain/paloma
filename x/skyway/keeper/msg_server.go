package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	utilkeeper "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/x/skyway/types"
	tokenfactorytypes "github.com/palomachain/paloma/v2/x/tokenfactory/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// SendToRemote handles MsgSendToRemote
func (k msgServer) SendToRemote(c context.Context, msg *types.MsgSendToRemote) (*types.MsgSendToRemoteResponse, error) {
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
	if k.InvalidSendToRemoteAddress(ctx, *dest, *erc20) {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "destination address is invalid")
	}
	txID, err := k.AddToOutgoingPool(ctx, sender, *dest, msg.Amount, msg.GetChainReferenceId())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not add to outgoing pool")
	}
	return &types.MsgSendToRemoteResponse{}, ctx.EventManager().EmitTypedEvent(
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

	ci, err := k.EVMKeeper.GetChainInfo(ctx, batch.ChainReferenceID)
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
	valAddr, err := k.valAddressFromAccAddress(ctx, orchestrator)
	if err != nil {
		return err
	}
	// return an error if the validator isn't in the active set
	val, err := k.StakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}
	if val == nil || !val.IsBonded() {
		return sdkerrors.Wrap(errorsmod.ErrorInvalidSigner, "validator not in active set")
	}

	return nil
}

func (k msgServer) valAddressFromAccAddress(ctx context.Context, address string) (sdk.ValAddress, error) {
	orchaddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}
	validator, found, err := k.GetOrchestratorValidator(ctx, orchaddr)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	return utilkeeper.ValAddressFromBech32(k.AddressCodec, validator.GetOperator())
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
			AttestationId: fmt.Sprintf("%x", types.GetAttestationKey(msg.GetSkywayNonce(), hash)),
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

func (k msgServer) BatchSendToRemoteClaim(c context.Context, msg *types.MsgBatchSendToRemoteClaim) (*types.MsgBatchSendToRemoteClaimResponse, error) {
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

	return &types.MsgBatchSendToRemoteClaimResponse{}, nil
}

// Performs additional checks on msg to determine if it is valid
func additionalPatchChecks(ctx context.Context, k msgServer, msg *types.MsgBatchSendToRemoteClaim) error {
	contractAddress, err := types.NewEthAddress(msg.TokenContract)
	if err != nil {
		return sdkerrors.Wrap(err, "Invalid TokenContract on MsgBatchSendToRemoteClaim")
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

func (k msgServer) CancelSendToRemote(c context.Context, msg *types.MsgCancelSendToRemote) (*types.MsgCancelSendToRemoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}
	err = k.RemoveFromOutgoingPoolAndRefund(ctx, msg.TransactionId, sender)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelSendToRemoteResponse{}, nil
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
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) LightNodeSaleClaim(
	c context.Context,
	msg *types.MsgLightNodeSaleClaim,
) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchestrator validator inset")
	}

	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check Any value")
	}

	err = k.claimHandlerCommon(ctx, msgAny, msg)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (k msgServer) EstimateBatchGas(
	c context.Context,
	msg *types.MsgEstimateBatchGas,
) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(c)
	valAddr, err := k.valAddressFromAccAddress(ctx, msg.Metadata.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %w", err)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid MsgConfirmBatch")
	}
	contract, err := types.NewEthAddress(msg.TokenContract)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "eth address invalid")
	}

	if err := k.checkOrchestratorValidatorInSet(ctx, msg.Metadata.Creator); err != nil {
		return nil, fmt.Errorf("cannot accept esimtate from inactive validator: %w", err)
	}

	// fetch the outgoing batch given the nonce
	batch, err := k.GetOutgoingTXBatch(ctx, *contract, msg.Nonce)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find batch")
	}

	// check if we already have this estimate
	estimate, err := k.GetBatchGasEstimate(ctx, msg.Nonce, *contract, valAddr)
	if err != nil {
		return nil, err
	}
	if estimate != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "gas estimate already received from sender")
	}
	key, err := k.SetBatchGasEstimate(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, ctx.EventManager().EmitTypedEvent(
		&types.EventBatchConfirmKey{
			Message:         msg.Type(),
			BatchConfirmKey: string(key),
		},
	)
}

func (k *msgServer) OverrideNonceProposal(ctx context.Context, req *types.MsgNonceOverrideProposal) (*emptypb.Empty, error) {
	if req.Metadata.Creator != k.authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Metadata.Creator)
	}
	return &emptypb.Empty{}, k.overrideNonce(ctx, req.ChainReferenceId, req.Nonce)
}

func (k *msgServer) ReplenishLostGrainsProposal(ctx context.Context, req *types.MsgReplenishLostGrainsProposal) (*emptypb.Empty, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "message validation failed: %v", err)
	}
	if req.Metadata.Creator != k.authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Metadata.Creator)
	}

	ctx, commit := sdk.UnwrapSDKContext(ctx).CacheContext()
	store := k.GetStore(ctx, types.StoreModulePrefix)
	if store.Get(types.ReplenishedGrainRecordsKey) != nil {
		return nil, fmt.Errorf("operation may only be run once.")
	}

	for _, addr := range []string{
		"paloma17t5pd5l0d8a3p54r0p92src8tx2tvs7l2afmd9",
		"paloma1vlrsw0hf6ddkje99jtnzh4raaa4dqpwqapeaz6",
	} {

		amount := sdk.NewCoin("ugrain", math.NewInt(665625000000000))
		err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
		if err != nil {
			return nil, fmt.Errorf("MintCoins: %w", err)
		}

		receiver := sdk.MustAccAddressFromBech32(addr)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
			receiver,
			sdk.NewCoins(amount))
		if err != nil {
			return nil, fmt.Errorf("SendCoinsFromModuleToAccount: %w", err)
		}
	}

	store.Set(types.ReplenishedGrainRecordsKey, []byte("1"))

	commit()
	return &emptypb.Empty{}, nil
}

func (k *msgServer) SetERC20ToTokenDenom(ctx context.Context, msg *types.MsgSetERC20ToTokenDenom) (*emptypb.Empty, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sender, err := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "sender address invalid")
	}
	erc20, err := types.NewEthAddress(msg.Erc20)
	if err != nil || erc20 == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "erc20 address invalid")
	}
	ci, err := k.EVMKeeper.GetChainInfo(ctx, msg.ChainReferenceId)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrUnsupported, "sender address invalid")
	}
	_, _, err = tokenfactorytypes.DeconstructDenom(msg.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrUnsupported, "denom invalid")
	}
	md, err := k.tokenFactoryKeeper.GetAuthorityMetadata(ctx, msg.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrDenomNotFound, err.Error())
	}
	if md.Admin != sender.String() {
		return nil, sdkerrors.Wrap(types.ErrUnauthorized, "missing token admin permission")
	}

	if err := k.setDenomToERC20(ctx, ci.ChainReferenceID, msg.Denom, *erc20); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInternal, err.Error())
	}

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSetERC20ToTokenDenom,
			sdk.NewAttribute(types.AttributeKeyChainReferenceID, ci.ChainReferenceID),
			sdk.NewAttribute(types.AttributeKeyERC20Address, msg.Erc20),
			sdk.NewAttribute(types.AttributeKeyTokenDenom, msg.Denom),
		),
	})

	return &emptypb.Empty{}, nil
}

func (k *msgServer) SetERC20MappingProposal(ctx context.Context, req *types.MsgSetERC20MappingProposal) (*emptypb.Empty, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "message validation failed: %v", err)
	}
	if req.Authority != req.Metadata.Creator {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "creator mismatch; expected %s, got %s", k.authority, req.Metadata.Creator)
	}
	if req.Metadata.Creator != k.authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Metadata.Creator)
	}

	ctx, commit := sdk.UnwrapSDKContext(ctx).CacheContext()
	for _, m := range req.Mappings {
		ethAddr, err := types.NewEthAddress(m.GetErc20())
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrInvalid, "eth address %s: %v", m.GetErc20(), err)
		}
		if err := k.setDenomToERC20(ctx, m.GetChainReferenceId(), m.GetDenom(), *ethAddr); err != nil {
			return nil, err
		}
	}

	commit()
	return &emptypb.Empty{}, nil
}
