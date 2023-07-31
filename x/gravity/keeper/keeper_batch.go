package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/palomachain/paloma/x/gravity/types"
)

/////////////////////////////
/////// BATCH CONFIRMS     //
/////////////////////////////

// GetBatchConfirm returns a batch confirmation given its nonce, the token contract, and a validator address
func (k Keeper) GetBatchConfirm(ctx sdk.Context, nonce uint64, tokenContract types.EthAddress, validator sdk.AccAddress) *types.MsgConfirmBatch {
	store := ctx.KVStore(k.storeKey)
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		ctx.Logger().Error("invalid validator address")
		return nil
	}
	entity := store.Get(types.GetBatchConfirmKey(tokenContract, nonce, validator))
	if entity == nil {
		return nil
	}
	confirm := types.MsgConfirmBatch{
		Nonce:         nonce,
		TokenContract: tokenContract.GetAddress().Hex(),
		EthSigner:     "",
		Orchestrator:  "",
		Signature:     "",
	}
	k.cdc.MustUnmarshal(entity, &confirm)
	return &confirm
}

// SetBatchConfirm sets a batch confirmation by a validator
func (k Keeper) SetBatchConfirm(ctx sdk.Context, batch *types.MsgConfirmBatch) []byte {
	store := ctx.KVStore(k.storeKey)
	acc, err := sdk.AccAddressFromBech32(batch.Orchestrator)
	if err != nil {
		panic(sdkerrors.Wrap(err, "invalid Orchestrator address"))
	}
	contract, err := types.NewEthAddress(batch.TokenContract)
	if err != nil {
		panic(sdkerrors.Wrap(err, "invalid TokenContract"))
	}
	key := types.GetBatchConfirmKey(*contract, batch.Nonce, acc)
	store.Set(key, k.cdc.MustMarshal(batch))
	return key
}

// DeleteBatchConfirms deletes confirmations for an outgoing transaction batch
func (k Keeper) DeleteBatchConfirms(ctx sdk.Context, batch types.InternalOutgoingTxBatch) {
	store := ctx.KVStore(k.storeKey)
	for _, confirm := range k.GetBatchConfirmByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract) {
		orchestrator, err := sdk.AccAddressFromBech32(confirm.Orchestrator)
		if err == nil {
			confirmKey := types.GetBatchConfirmKey(batch.TokenContract, batch.BatchNonce, orchestrator)
			if store.Has(confirmKey) {
				store.Delete(confirmKey)
			}
		}
	}
}

// IterateBatchConfirmByNonceAndTokenContract iterates through all batch confirmations
// MARK finish-batches: this is where the key is iterated in the old (presumed working) code
// TODO: specify which nonce this is
func (k Keeper) IterateBatchConfirmByNonceAndTokenContract(ctx sdk.Context, nonce uint64, tokenContract types.EthAddress, cb func([]byte, types.MsgConfirmBatch) bool) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetBatchConfirmNonceContractPrefix(tokenContract, nonce)
	iter := store.Iterator(prefixRange([]byte(prefix)))

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
}

// GetBatchConfirmByNonceAndTokenContract returns the batch confirms
func (k Keeper) GetBatchConfirmByNonceAndTokenContract(ctx sdk.Context, nonce uint64, tokenContract types.EthAddress) (out []types.MsgConfirmBatch) {
	k.IterateBatchConfirmByNonceAndTokenContract(ctx, nonce, tokenContract, func(_ []byte, msg types.MsgConfirmBatch) bool {
		out = append(out, msg)
		return false
	})
	return
}

// IterateBatchConfirms iterates through all batch confirmations
func (k Keeper) IterateBatchConfirms(ctx sdk.Context, cb func([]byte, types.MsgConfirmBatch) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
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
