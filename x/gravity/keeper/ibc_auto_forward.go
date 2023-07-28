// Contains all the logic for IBC Auto-Forwarding from ethereum to foreign ibc-enabled cosmos chains
// This logic should be used by attestation_handler, msg_server, and the CLI methods
/*
Flow: On processing a SendToCosmos attestation in attestation handler with a foreign-prefixed CosmosReceiver address,
  a new entry is created in the PendingIbcAutoForwards queue. Later a MsgExecuteIbcAutoForwards can be submitted to
  clear the queue and move the funds to their destination chains over IBC.
This queue is necessary due to a Tendermint bug where ctx.EventManager().EmitEvent() has no effect when called from
  EndBlocker. The queue allows processing SendToCosmos attestations from EndBlocker while emitting events from DeliverTx.
*/

package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	bech32ibctypes "github.com/palomachain/paloma/x/bech32ibc/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

// ValidatePendingIbcAutoForward performs basic validation, asserts the nonce is not ahead of what gravity is aware of,
// requires ForeignReceiver's bech32 prefix to be registered and match with IbcChannel, and gravity module must have the
// funds to meet this forward amount
func (k Keeper) ValidatePendingIbcAutoForward(ctx sdk.Context, forward types.PendingIbcAutoForward) error {
	if err := forward.ValidateBasic(); err != nil {
		return err
	}

	latestEventNonce := k.GetLastObservedEventNonce(ctx)
	if forward.EventNonce > latestEventNonce {
		return sdkerrors.Wrap(types.ErrInvalid, "EventNonce must be <= latest observed event nonce")
	}
	prefix, _, err := bech32.DecodeAndConvert(forward.ForeignReceiver)
	if err != nil { // Covered by ValidateBasic, but check anyway to avoid linter issues
		return sdkerrors.Wrapf(err, "ForeignReceiver %s is not a valid bech32 address", forward.ForeignReceiver)
	}
	hrpRecord, err := k.bech32IbcKeeper.GetHrpIbcRecord(ctx, prefix)
	if err != nil {
		return sdkerrors.Wrapf(bech32ibctypes.ErrInvalidHRP, "ForeignReciever %s has an invalid or unregistered prefix", forward.ForeignReceiver)
	}
	if forward.IbcChannel != hrpRecord.SourceChannel {
		return sdkerrors.Wrapf(types.ErrMismatched, "IbcChannel %s does not match the registered prefix's IBC channel %v",
			forward.IbcChannel, hrpRecord.String(),
		)
	}
	modAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress()
	modBal := k.bankKeeper.GetBalance(ctx, modAcc, forward.Token.Denom)
	if modBal.IsLT(*forward.Token) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "Gravity Module account does not have enough funds (%s) for a forward of %s",
			modBal.String(), forward.Token.String(),
		)
	}

	return nil
}

// GetNextPendingIbcAutoForward returns the first pending IBC Auto-Forward in the queue
func (k Keeper) GetNextPendingIbcAutoForward(ctx sdk.Context) *types.PendingIbcAutoForward {
	store := ctx.KVStore(k.storeKey)
	prefix := types.PendingIbcAutoForwards
	iter := store.Iterator(prefixRange(prefix))
	defer iter.Close()
	if iter.Valid() {
		forward := types.PendingIbcAutoForward{
			ForeignReceiver: "",
			Token:           nil,
			IbcChannel:      "",
			EventNonce:      0,
		}
		k.cdc.MustUnmarshal(iter.Value(), &forward)

		return &forward
	}
	return nil
}

// PendingIbcAutoForwards returns an ordered slice of the queued IBC Auto-Forward sends to IBC-enabled chains
func (k Keeper) PendingIbcAutoForwards(ctx sdk.Context, limit uint64) []*types.PendingIbcAutoForward {
	forwards := make([]*types.PendingIbcAutoForward, 0)

	k.IteratePendingIbcAutoForwards(ctx, func(key []byte, forward *types.PendingIbcAutoForward) (stop bool) {
		forwards = append(forwards, forward)
		if limit != 0 && uint64(len(forwards)) >= limit {
			return true
		}
		return false
	})

	return forwards
}

// IteratePendingIbcAutoForwards executes the given callback on each PendingIbcAutoForward in the store
// cb should return true to stop iteration, false to continue
func (k Keeper) IteratePendingIbcAutoForwards(ctx sdk.Context, cb func(key []byte, forward *types.PendingIbcAutoForward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.PendingIbcAutoForwards)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		forward := new(types.PendingIbcAutoForward)
		k.cdc.MustUnmarshal(iter.Value(), forward)

		if cb(iter.Key(), forward) {
			break
		}
	}
}

// addPendingIbcAutoForward enqueues a single new pending IBC Auto-Forward send to an IBC-enabled chain
// IbcAutoForwards are added to
func (k Keeper) addPendingIbcAutoForward(ctx sdk.Context, forward types.PendingIbcAutoForward, token string) error {
	if err := k.ValidatePendingIbcAutoForward(ctx, forward); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	key := types.GetPendingIbcAutoForwardKey(forward.EventNonce)

	if store.Has(key) {
		return sdkerrors.Wrapf(types.ErrDuplicate,
			"Pending IBC Auto-Forward Queue already has an entry with nonce %v", forward.EventNonce,
		)
	}
	store.Set(key, k.cdc.MustMarshal(&forward))

	k.Logger(ctx).Info("SendToCosmos Pending IBC Auto-Forward", "ibcReceiver", forward.ForeignReceiver,
		"token", token, "denom", forward.Token.Denom, "amount", forward.Token.Amount.String(),
		"ibc-port", k.ibcTransferKeeper.GetPort(ctx), "ibcChannel", forward.IbcChannel, "claimNonce", forward.EventNonce,
		"cosmosBlockTime", ctx.BlockTime(), "cosmosBlockHeight", ctx.BlockHeight(),
	)

	return ctx.EventManager().EmitTypedEvent(&types.EventSendToCosmosPendingIbcAutoForward{
		Nonce:    fmt.Sprint(forward.EventNonce),
		Receiver: forward.ForeignReceiver,
		Token:    token,
		Amount:   forward.Token.Amount.String(),
		Channel:  forward.IbcChannel,
	})
}

// deletePendingIbcAutoForward removes a single pending IBC Auto-Forward send to an IBC-enabled chain from the store
// WARNING: this should only be called while clearing the queue in ClearNextPendingIbcAutoForward
func (k Keeper) deletePendingIbcAutoForward(ctx sdk.Context, eventNonce uint64) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPendingIbcAutoForwardKey(eventNonce)
	if !store.Has(key) {
		return sdkerrors.Wrapf(types.ErrInvalid, "No PendingIbcAutoForward with nonce %v in the store", eventNonce)
	}
	store.Delete(key)
	return nil
}

// ProcessPendingIbcAutoForwards processes and dequeues many pending IBC Auto-Forwards, either sending the funds to their
// respective destination chains or on error sending the funds to the local gravity-prefixed account
// See ProcessNextPendingIbcAutoForward for more details
func (k Keeper) ProcessPendingIbcAutoForwards(ctx sdk.Context, forwardsToClear uint64) error {
	for i := uint64(0); i < forwardsToClear; i++ {
		stop, err := k.ProcessNextPendingIbcAutoForward(ctx)
		if err != nil {
			return sdkerrors.Wrapf(err, "unable to process Pending IBC Auto-Forward number %v", i)
		}
		if stop {
			break
		}
	}

	return nil
}

// ProcessNextPendingIbcAutoForward processes and dequeues a single pending IBC Auto-Forward, initially sending the funds
// to the local gravity-prefixed account and then initiating an ibc transfer to the destination chain
// e.g. if the SendToCosmos CosmosReceiver was [cosmos1|ADDR|COSMOSCHECKSUM] the gravity-prefixed account will be
// [gravity1|ADDR|GRAVITYCHECKSUM] instead, the same account will retain control of the funds but they will need to use
// the gravity chain directly.
// Return: stop - true if there are no items to process
// Return: err  - nil if successful, non-nil if there was an issue sending funds to the user
func (k Keeper) ProcessNextPendingIbcAutoForward(ctx sdk.Context) (stop bool, err error) {
	forward := k.GetNextPendingIbcAutoForward(ctx)
	if forward == nil {
		return true, nil // No forwards to process, exit early
	}
	if err := forward.ValidateBasic(); err != nil { // double-check the forward before sending it
		// Fail this tx
		panic(fmt.Sprintf("Invalid forward found in Pending IBC Auto-Forward queue: %s", err.Error()))
	}
	// Point of no return: the funds will be sent somewhere, either the IBC address, local address or the community pool
	err = k.deletePendingIbcAutoForward(ctx, forward.EventNonce)
	if err != nil {
		// Fail this tx
		panic(fmt.Sprintf("Discovered nonexistent Pending IBC Auto-Forward in the queue %s", forward.String()))
	}

	portId := k.ibcTransferKeeper.GetPort(ctx)

	// This local gravity user receives the coins if the ibc transaction fails
	var fallback sdk.AccAddress
	fallback, err = types.IBCAddressFromBech32(forward.ForeignReceiver)
	if err != nil {
		panic(fmt.Sprintf("Invalid ForeignReceiver found in Pending IBC Auto-Forward queue: %s [[%+v]]", err.Error(), forward))
	}

	coins := sdk.NewCoins(*forward.Token)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, fallback, coins)
	if err != nil {
		// Couldn't send to fallback account, need to try community pool
		return false, k.SendToCommunityPool(ctx, coins)
	}

	timeoutTime := thirtyDaysInFuture(ctx) // Set the ibc transfer to expire ~one month from now

	msgTransfer := createIbcMsgTransfer(portId, *forward, fallback.String(), uint64(timeoutTime.UnixNano()))

	// Make the ibc-transfer attempt
	wCtx := sdk.WrapSDKContext(ctx)
	_, recoverableErr := k.ibcTransferKeeper.Transfer(wCtx, &msgTransfer)
	ctx = sdk.UnwrapSDKContext(wCtx)

	// Log + emit event
	if recoverableErr == nil {
		k.logEmitIbcForwardSuccessEvent(ctx, *forward, msgTransfer)
	} else {
		// Funds have already been sent to the fallback user, emit a failure log
		/*
			k.ibcTransferKeeper.Transfer() failure cases (and resolution)

			1. Sender is invalid bech32 (checked before send)
			2. IBC Sends are not enabled (local receiver)
			3. Could not find the channel end from portId + sourceChannel (checked before send)
			4. Could not find the next sequence for the source end (local receiver)
			5. ibc-transfer module doesn't own the capability for the source end (local receiver)
			6. Could not get an ibc denom's full path from its hash (local receiver of invalid token?)
			7. Unable to send a gravity-native token to the ibc-transfer escrow address (local receiver)
			8. Unable to send ibc vouchers to the ibc-transfer module, or sent to ibc-transfer module but couldn't burn
				(local receiver of a token which can't be sent?)
			9. Could not send packet to the channel e.g. connection issues, misconfigured packet, timeouts, sequences
			    (local receiver)
		*/
		k.logEmitIbcForwardFailureEvent(ctx, *forward, recoverableErr)
	}
	return false, nil // Error case has been handled, funds are in receiver's control locally or on IBC chain
}

// createIbcMsgTransfer creates a MsgTransfer for the given pending `forward` on port `portId` sent from `sender`,
// with the given timeout timestamp and a zero timeout block height
func createIbcMsgTransfer(portId string, forward types.PendingIbcAutoForward, sender string, timeoutTimestampNs uint64) ibctransfertypes.MsgTransfer {
	zeroHeight := ibcclienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: 0,
	}
	return *ibctransfertypes.NewMsgTransfer(
		portId,
		forward.IbcChannel,
		*forward.Token,
		sender,
		forward.ForeignReceiver,
		zeroHeight, // Do not use block height based timeout
		timeoutTimestampNs,
		"",
	)
}

// thirtyDaysInFuture creates a time.Time exactly 30 days from the last BlockTime for use in createIbcMsgTransfer
func thirtyDaysInFuture(ctx sdk.Context) time.Time {
	approxNow := ctx.BlockTime()
	// Get the offset from zero of 30 days in the future
	return approxNow.Add(time.Hour * 24 * 30)
}

// logEmitIbcForwardSuccessEvent logs for successful IBC Auto-Forwarding and emits a
// EventSendToCosmosExecutedIbcAutoForward type event
func (k Keeper) logEmitIbcForwardSuccessEvent(
	ctx sdk.Context,
	forward types.PendingIbcAutoForward,
	msgTransfer ibctransfertypes.MsgTransfer,
) {
	k.Logger(ctx).Info("SendToCosmos IBC Auto-Forward", "ibcReceiver", forward.ForeignReceiver, "denom", forward.Token.Denom,
		"amount", forward.Token.Amount.String(), "ibc-port", msgTransfer.SourcePort, "ibcChannel", forward.IbcChannel,
		"timeoutHeight", msgTransfer.TimeoutHeight.String(), "timeoutTimestamp", msgTransfer.TimeoutTimestamp,
		"claimNonce", forward.EventNonce, "cosmosBlockHeight", ctx.BlockHeight(),
	)

	err := ctx.EventManager().EmitTypedEvent(&types.EventSendToCosmosExecutedIbcAutoForward{
		Nonce:         fmt.Sprint(forward.EventNonce),
		Receiver:      forward.ForeignReceiver,
		Token:         forward.Token.Denom,
		Amount:        forward.Token.Amount.String(),
		Channel:       forward.IbcChannel,
		TimeoutHeight: msgTransfer.TimeoutHeight.String(),
		TimeoutTime:   fmt.Sprint(msgTransfer.TimeoutTimestamp),
	})
	if err != nil {
		panic(err)
	}
}

// logEmitIbcForwardFailureEvent logs failed IBC Auto-Forwarding and emits a EventSendToCosmosLocal type event
func (k Keeper) logEmitIbcForwardFailureEvent(ctx sdk.Context, forward types.PendingIbcAutoForward, err error) {
	var localReceiver sdk.AccAddress
	localReceiver, er := types.IBCAddressFromBech32(forward.ForeignReceiver) // checked valid bech32 receiver earlier
	if er != nil {
		panic(err)
	}
	k.Logger(ctx).Error("SendToCosmos IBC Auto-Forward Failure: funds sent to local address",
		"localReceiver", localReceiver, "denom", forward.Token.Denom, "amount", forward.Token.Amount.String(),
		"failedIbcPort", ibctransfertypes.PortID, "failedIbcChannel", forward.IbcChannel,
		"claimNonce", forward.EventNonce, "cosmosBlockHeight", ctx.BlockHeight(), "err", err,
	)

	er = ctx.EventManager().EmitTypedEvent(&types.EventSendToCosmosLocal{
		Nonce:    fmt.Sprint(forward.EventNonce),
		Receiver: forward.ForeignReceiver,
		Token:    forward.Token.Denom,
		Amount:   forward.Token.Amount.String(),
	})
	if er != nil {
		panic(err)
	}
}
