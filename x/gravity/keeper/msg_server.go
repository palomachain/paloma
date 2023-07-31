package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

// BasisPointDivisor used in determining if a SendToEth fee meets the governance-controlled minimum
const BasisPointDivisor uint64 = 10000

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// SetOrchestratorAddress handles MsgSetOrchestratorAddress
func (k msgServer) SetOrchestratorAddress(c context.Context, msg *types.MsgSetOrchestratorAddress) (*types.MsgSetOrchestratorAddressResponse, error) {
	// ensure that this passes validation, checks the key validity
	err := msg.ValidateBasic()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Key not valid")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// check the following, all should be validated in validate basic
	// check the following, all should be validated in validate basic
	val, e1 := sdk.ValAddressFromBech32(msg.Validator)
	orch, e2 := sdk.AccAddressFromBech32(msg.Orchestrator)
	ethAddr, e3 := types.NewEthAddress(msg.EthAddress)
	if e1 != nil || e2 != nil || e3 != nil {
		return nil, sdkerrors.Wrap(err, "Key not valid")
	}

	// ensure that the validator exists
	if k.Keeper.StakingKeeper.Validator(ctx, val) == nil {
		return nil, sdkerrors.Wrap(stakingtypes.ErrNoValidatorFound, val.String())
	}

	_, foundExistingOrchestratorKey := k.GetOrchestratorValidator(ctx, orch)
	_, foundExistingEthAddress := k.GetEthAddressByValidator(ctx, val)

	// ensure that the validator does not have an existing key
	if foundExistingOrchestratorKey || foundExistingEthAddress {
		return nil, sdkerrors.Wrap(types.ErrResetDelegateKeys, val.String())
	}

	// ensure that neither key is a duplicate
	delegateKeys := k.GetDelegateKeys(ctx)
	for i := range delegateKeys {
		if delegateKeys[i].EthAddress == ethAddr.GetAddress().Hex() {
			return nil, types.ErrDuplicateEthereumKey
		}
		if delegateKeys[i].Orchestrator == orch.String() {
			return nil, types.ErrDuplicateOrchestratorKey
		}
	}

	// set the orchestrator address and the ethereum address
	k.SetOrchestratorValidator(ctx, val, orch)
	k.SetEthAddressForValidator(ctx, val, *ethAddr)

	return &types.MsgSetOrchestratorAddressResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventSetOperatorAddress{
			Message: msg.Type(),
			Address: orch.String(),
		},
	)
}

// ValsetConfirm handles MsgValsetConfirm
func (k msgServer) ValsetConfirm(c context.Context, msg *types.MsgValsetConfirm) (*types.MsgValsetConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	valset := k.GetValset(ctx, msg.Nonce) // A valset request was previously created
	if valset == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find valset")
	}

	gravityID := k.GetGravityID(ctx)
	checkpoint := valset.GetCheckpoint(gravityID)
	orchaddr, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}
	err = k.confirmHandlerCommon(ctx, msg.EthAddress, orchaddr, msg.Signature, checkpoint)
	if err != nil {
		return nil, err
	}

	// persist signature
	if k.GetValsetConfirm(ctx, msg.Nonce, orchaddr) != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "signature duplicate")
	}
	key := k.SetValsetConfirm(ctx, *msg)

	return &types.MsgValsetConfirmResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventValsetConfirmKey{
			Message: msg.Type(),
			Key:     string(key),
		},
	)
}

// SendToEth handles MsgSendToEth
func (k msgServer) SendToEth(c context.Context, msg *types.MsgSendToEth) (*types.MsgSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid sender")
	}

	dest, err := types.NewEthAddress(msg.EthDest)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid eth dest")
	}

	_, erc20, err := k.DenomToERC20Lookup(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid denom")
	}

	if k.InvalidSendToEthAddress(ctx, *dest, *erc20) {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "destination address is invalid or blacklisted")
	}

	// Collect the ChainFee and give to stakers, ensuring it meets MinChainFeeBasisPoints
	if err := k.checkAndDeductSendToEthFees(ctx, sender, msg.Amount, msg.ChainFee); err != nil {
		return nil, sdkerrors.Wrapf(err, "Could not deduct chainFee %v from account %v", msg.ChainFee.String(), msg.Sender)
	}

	txID, err := k.AddToOutgoingPool(ctx, sender, *dest, msg.Amount, msg.BridgeFee)
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

// checkAndDeductSendToEthFees asserts that the minimum chainFee has been met for the given sendAmount
func (k msgServer) checkAndDeductSendToEthFees(ctx sdk.Context, sender sdk.AccAddress, sendAmount sdk.Coin, chainFee sdk.Coin) error {
	// Compute the minimum fee which must be paid
	minFeeBasisPoints := int64(0)
	params, err := k.Keeper.GetParamsIfSet(ctx)
	if err == nil {
		// The params have been set, get the min send to eth fee
		minFeeBasisPoints = int64(params.MinChainFeeBasisPoints)
	}
	minFee := sdk.NewDecFromInt(sendAmount.Amount).
		QuoInt64(int64(BasisPointDivisor)).
		MulInt64(minFeeBasisPoints).
		TruncateInt()

	// Require that the minimum has been met
	if minFee.GT(sdk.ZeroInt()) { // Ignore fees too low to collect
		minFeeCoin := sdk.NewCoin(sendAmount.GetDenom(), minFee)
		if chainFee.IsLT(minFeeCoin) {
			err := sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFee,
				"chain fee provided [%s] is insufficient, need at least [%s]",
				chainFee,
				minFeeCoin,
			)
			return err
		}
	}

	// Finally, collect any provided fees
	// nolint: exhaustruct
	if !(chainFee == sdk.Coin{}) && chainFee.Amount.IsPositive() {
		senderAcc := k.accountKeeper.GetAccount(ctx, sender)

		err = sdkante.DeductFees(k.bankKeeper, ctx, senderAcc, sdk.NewCoins(chainFee))
		if err != nil {
			ctx.Logger().Error("Could not deduct MsgSendToEth fee!", "error", err, "account", senderAcc, "chainFee", chainFee)
			return err
		}

		// Report the fee collection to the event log
		return ctx.EventManager().EmitTypedEvent(
			&types.EventSendToEthFeeCollected{
				Sender:     sender.String(),
				SendAmount: sendAmount.String(),
				FeeAmount:  chainFee.String(),
			},
		)
	}

	return nil
}

// RequestBatch handles MsgRequestBatch
func (k msgServer) RequestBatch(c context.Context, msg *types.MsgRequestBatch) (*types.MsgRequestBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Check if the denom is a gravity coin, if not, check if there is a deployed ERC20 representing it.
	// If not, error out
	_, tokenContract, err := k.DenomToERC20Lookup(ctx, msg.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not look up erc 20 denominator")
	}

	batch, err := k.BuildOutgoingTXBatch(ctx, *tokenContract, OutgoingTxBatchSize)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not build outgoing tx batch")
	}

	return &types.MsgRequestBatchResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventBatchCreated{
			Message:    msg.Type(),
			BatchNonce: fmt.Sprint(batch.BatchNonce),
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
	batch := k.GetOutgoingTXBatch(ctx, *contract, msg.Nonce)
	if batch == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find batch")
	}

	gravityID := k.GetGravityID(ctx)
	checkpoint := batch.GetCheckpoint(gravityID)
	orchaddr, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}

	err = k.confirmHandlerCommon(ctx, msg.EthSigner, orchaddr, msg.Signature, checkpoint)
	if err != nil {
		return nil, err
	}

	// check if we already have this confirm
	if k.GetBatchConfirm(ctx, msg.Nonce, *contract, orchaddr) != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "duplicate signature")
	}
	key := k.SetBatchConfirm(ctx, msg)

	return &types.MsgConfirmBatchResponse{}, ctx.EventManager().EmitTypedEvent(
		&types.EventBatchConfirmKey{
			Message:         msg.Type(),
			BatchConfirmKey: string(key),
		},
	)
}

// ConfirmLogicCall handles MsgConfirmLogicCall
func (k msgServer) ConfirmLogicCall(c context.Context, msg *types.MsgConfirmLogicCall) (*types.MsgConfirmLogicCallResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	invalidationIdBytes, err := hex.DecodeString(msg.InvalidationId)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "invalidation id encoding")
	}

	// fetch the outgoing logic given the nonce
	logic := k.GetOutgoingLogicCall(ctx, invalidationIdBytes, msg.InvalidationNonce)
	if logic == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find logic")
	}

	gravityID := k.GetGravityID(ctx)
	checkpoint := logic.GetCheckpoint(gravityID)
	orchaddr, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}
	err = k.confirmHandlerCommon(ctx, msg.EthSigner, orchaddr, msg.Signature, checkpoint)
	if err != nil {
		return nil, err
	}

	// check if we already have this confirm
	if k.GetLogicCallConfirm(ctx, invalidationIdBytes, msg.InvalidationNonce, orchaddr) != nil {
		return nil, sdkerrors.Wrap(types.ErrDuplicate, "duplicate signature")
	}

	k.SetLogicCallConfirm(ctx, msg)

	return nil, nil
}

// checkOrchestratorValidatorInSet checks that the orchestrator refers to a validator that is
// currently in the set
func (k msgServer) checkOrchestratorValidatorInSet(ctx sdk.Context, orchestrator string) error {
	orchaddr, err := sdk.AccAddressFromBech32(orchestrator)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "acc address invalid")
	}
	validator, found := k.GetOrchestratorValidator(ctx, orchaddr)
	if !found {
		return sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	// return an error if the validator isn't in the active set
	val := k.StakingKeeper.Validator(ctx, validator.GetOperator())
	if val == nil || !val.IsBonded() {
		return sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, "validator not in active set")
	}

	return nil
}

// claimHandlerCommon is an internal function that provides common code for processing claims once they are
// translated from the message to the Ethereum claim interface
func (k msgServer) claimHandlerCommon(ctx sdk.Context, msgAny *codectypes.Any, msg types.EthereumClaim) error {
	// Add the claim to the store
	_, err := k.Attest(ctx, msg, msgAny)
	if err != nil {
		return sdkerrors.Wrap(err, "create attestation")
	}
	hash, err := msg.ClaimHash()
	if err != nil {
		return sdkerrors.Wrap(err, "unable to compute claim hash")
	}

	// Emit the handle message event
	return ctx.EventManager().EmitTypedEvent(
		&types.EventClaim{
			Message:       string(msg.GetType()),
			ClaimHash:     string(hash),
			AttestationId: string(types.GetAttestationKey(msg.GetEventNonce(), hash)),
		},
	)
}

// confirmHandlerCommon is an internal function that provides common code for processing claim messages
func (k msgServer) confirmHandlerCommon(ctx sdk.Context, ethAddress string, orchestrator sdk.AccAddress, signature string, checkpoint []byte) error {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "signature decoding")
	}

	submittedEthAddress, err := types.NewEthAddress(ethAddress)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "invalid eth address")
	}

	validator, found := k.GetOrchestratorValidator(ctx, orchestrator)
	if !found {
		return sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	if !validator.IsBonded() && !validator.IsUnbonding() {
		// We must only accept confirms from bonded or unbonding validators
		return sdkerrors.Wrap(types.ErrInvalid, "validator is unbonded")
	}

	if err := sdk.VerifyAddressFormat(validator.GetOperator()); err != nil {
		return sdkerrors.Wrapf(err, "discovered invalid validator address for orchestrator %v", orchestrator)
	}

	ethAddressFromStore, found := k.GetEthAddressByValidator(ctx, validator.GetOperator())
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

// DepositClaim handles MsgSendToCosmosClaim
// TODO it is possible to submit an old msgDepositClaim (old defined as covering an event nonce that has already been
// executed aka 'observed' and had it's slashing window expire) that will never be cleaned up in the endblocker. This
// should not be a security risk as 'old' events can never execute but it does store spam in the chain.
func (k msgServer) SendToCosmosClaim(c context.Context, msg *types.MsgSendToCosmosClaim) (*types.MsgSendToCosmosClaimResponse, error) {
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

	return &types.MsgSendToCosmosClaimResponse{}, nil
}

// ExecuteIbcAutoForwards moves pending IBC Auto-Forwards to their respective chains by calling ibc-transfer's Transfer
// function with all the relevant information
// Note: this endpoint and the related queue are necessary due to a Tendermint bug where events created in EndBlocker
// do not appear. We process SendToCosmos observations in EndBlocker but are therefore unable to auto-forward these txs
// in the same block. This endpoint triggers the creation of those ibc-transfer events which relayers watch for.
func (k msgServer) ExecuteIbcAutoForwards(c context.Context, msg *types.MsgExecuteIbcAutoForwards) (*types.MsgExecuteIbcAutoForwardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := k.ProcessPendingIbcAutoForwards(ctx, msg.GetForwardsToClear()); err != nil {
		return nil, err
	}

	return &types.MsgExecuteIbcAutoForwardsResponse{}, nil
}

// WithdrawClaim handles MsgBatchSendToEthClaim
// TODO it is possible to submit an old msgWithdrawClaim (old defined as covering an event nonce that has already been
// executed aka 'observed' and had it's slashing window expire) that will never be cleaned up in the endblocker. This
// should not be a security risk as 'old' events can never execute but it does store spam in the chain.
func (k msgServer) BatchSendToEthClaim(c context.Context, msg *types.MsgBatchSendToEthClaim) (*types.MsgBatchSendToEthClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchestrator validator")
	}

	/* Perform some additional checks on the input to determine if it is valid before allowing it on the chain
	   Note that because of the gas meter we must avoid calls which consume gas, like fetching data from the keeper
	*/
	additionalPatchChecks(ctx, k, msg)

	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(sdkerrors.Wrap(err, "Could not check Any value"))
	}

	err = k.claimHandlerCommon(ctx, msgAny, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgBatchSendToEthClaimResponse{}, nil
}

// Performs additional checks on msg to determine if it is valid
func additionalPatchChecks(ctx sdk.Context, k msgServer, msg *types.MsgBatchSendToEthClaim) {
	contractAddress, err := types.NewEthAddress(msg.TokenContract)

	if err != nil {
		panic(sdkerrors.Wrap(err, "Invalid TokenContract on MsgBatchSendToEthClaim"))
	}

	// Replicate the following but without using a gas meter:
	b := k.GetOutgoingTXBatch(ctx, *contractAddress, msg.BatchNonce)
	if b == nil {
		// Batch deleted, just add the vote to the stored attestation
		return
	}

	if b.BatchTimeout <= msg.EthBlockHeight {
		panic(fmt.Errorf("batch with nonce %d submitted after it timed out (submission %d >= timeout %d)", msg.BatchNonce, msg.EthBlockHeight, b.BatchTimeout))
	}
}

// ERC20Deployed handles MsgERC20Deployed
func (k msgServer) ERC20DeployedClaim(c context.Context, msg *types.MsgERC20DeployedClaim) (*types.MsgERC20DeployedClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchestrator validator in set")
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check Any value")
	}
	err = k.claimHandlerCommon(ctx, any, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgERC20DeployedClaimResponse{}, nil
}

// LogicCallExecutedClaim handles claims for executing a logic call on Ethereum
func (k msgServer) LogicCallExecutedClaim(c context.Context, msg *types.MsgLogicCallExecutedClaim) (*types.MsgLogicCallExecutedClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchestrator validator in set")
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check Any value")
	}
	err = k.claimHandlerCommon(ctx, any, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgLogicCallExecutedClaimResponse{}, nil
}

// ValsetUpdatedClaim handles claims for executing a validator set update on Ethereum
func (k msgServer) ValsetUpdateClaim(c context.Context, msg *types.MsgValsetUpdatedClaim) (*types.MsgValsetUpdatedClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.checkOrchestratorValidatorInSet(ctx, msg.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check orchestrator validator in set")
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Could not check Any value")
	}
	err = k.claimHandlerCommon(ctx, any, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgValsetUpdatedClaimResponse{}, nil
}

func (k msgServer) CancelSendToEth(c context.Context, msg *types.MsgCancelSendToEth) (*types.MsgCancelSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
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

	err := k.CheckBadSignatureEvidence(ctx, msg)
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
