package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/gravity/types"
)

/////////////////////////////
/////// BATCH CONFIRMS     //
/////////////////////////////

// GetBatchConfirm returns a batch confirmation given its nonce, the token contract, and a validator address
func (k Keeper) GetBatchConfirm(ctx context.Context, nonce uint64, tokenContract types.EthAddress, validator sdk.AccAddress) (*types.MsgConfirmBatch, error) {
	store := k.GetStore(ctx)
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).Error("invalid validator address")
		return nil, nil
	}
	batchConfirmKey, err := types.GetBatchConfirmKey(tokenContract, nonce, validator)
	if err != nil {
		return nil, err
	}
	entity := store.Get(batchConfirmKey)
	if entity == nil {
		return nil, nil
	}
	confirm := types.MsgConfirmBatch{
		Nonce:         nonce,
		TokenContract: tokenContract.GetAddress().Hex(),
		EthSigner:     "",
		Orchestrator:  "",
		Signature:     "",
	}
	k.cdc.MustUnmarshal(entity, &confirm)
	return &confirm, nil
}

// SetBatchConfirm sets a batch confirmation by a validator
func (k Keeper) SetBatchConfirm(ctx context.Context, batch *types.MsgConfirmBatch) ([]byte, error) {
	store := k.GetStore(ctx)
	acc, err := sdk.AccAddressFromBech32(batch.Orchestrator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid Orchestrator address")
	}
	contract, err := types.NewEthAddress(batch.TokenContract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid TokenContract")
	}
	key, err := types.GetBatchConfirmKey(*contract, batch.Nonce, acc)
	if err != nil {
		return nil, err
	}
	store.Set(key, k.cdc.MustMarshal(batch))
	return key, nil
}

// DeleteBatchConfirms deletes confirmations for an outgoing transaction batch
func (k Keeper) DeleteBatchConfirms(ctx context.Context, batch types.InternalOutgoingTxBatch) error {
	store := k.GetStore(ctx)
	batchConfirms, err := k.GetBatchConfirmByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract)
	if err != nil {
		return err
	}
	for _, confirm := range batchConfirms {
		orchestrator, err := sdk.AccAddressFromBech32(confirm.Orchestrator)
		if err != nil {
			return err
		}
		confirmKey, err := types.GetBatchConfirmKey(batch.TokenContract, batch.BatchNonce, orchestrator)
		if err != nil {
			return err
		}
		if store.Has(confirmKey) {
			store.Delete(confirmKey)
		}
	}
	return nil
}

// IterateBatchConfirmByNonceAndTokenContract iterates through all batch confirmations
// MARK finish-batches: this is where the key is iterated in the old (presumed working) code
// TODO: specify which nonce this is
func (k Keeper) IterateBatchConfirmByNonceAndTokenContract(ctx context.Context, nonce uint64, tokenContract types.EthAddress, cb func([]byte, types.MsgConfirmBatch) bool) error {
	store := k.GetStore(ctx)
	prefix := types.GetBatchConfirmNonceContractPrefix(tokenContract, nonce)
	start, end, err := prefixRange(prefix)
	if err != nil {
		return err
	}
	iter := store.Iterator(start, end)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		confirm := types.MsgConfirmBatch{
			Nonce:         nonce,
			TokenContract: tokenContract.GetAddress().Hex(),
			EthSigner:     "",
			Orchestrator:  "",
			Signature:     "",
		}
		k.cdc.MustUnmarshal(iter.Value(), &confirm)
		// cb returns true to stop early
		if cb(iter.Key(), confirm) {
			break
		}
	}
	return nil
}

// GetBatchConfirmByNonceAndTokenContract returns the batch confirms
func (k Keeper) GetBatchConfirmByNonceAndTokenContract(ctx context.Context, nonce uint64, tokenContract types.EthAddress) (out []types.MsgConfirmBatch, err error) {
	err = k.IterateBatchConfirmByNonceAndTokenContract(ctx, nonce, tokenContract, func(_ []byte, msg types.MsgConfirmBatch) bool {
		out = append(out, msg)
		return false
	})
	return
}

// IterateBatchConfirms iterates through all batch confirmations
func (k Keeper) IterateBatchConfirms(ctx context.Context, cb func([]byte, types.MsgConfirmBatch) (stop bool)) {
	store := k.GetStore(ctx)
	prefixStore := prefix.NewStore(store, types.BatchConfirmKey)
	iter := prefixStore.Iterator(nil, nil)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var confirm types.MsgConfirmBatch
		k.cdc.MustUnmarshal(iter.Value(), &confirm)

		// cb returns true to stop early
		if cb(iter.Key(), confirm) {
			break
		}
	}
}
