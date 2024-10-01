package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/skyway/types"
)

func (k Keeper) GetBatchGasEstimate(ctx context.Context, nonce uint64, tokenContract types.EthAddress, validator sdk.ValAddress) (*types.MsgEstimateBatchGas, error) {
	store := k.GetStore(ctx, types.StoreModulePrefix)
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		liblog.FromKeeper(ctx, k).WithError(err).Error("invalid validator address")
		return nil, nil
	}
	batchGasEstimateKey, err := types.GetBatchGasEstimateKey(tokenContract, nonce, validator)
	if err != nil {
		return nil, err
	}
	entity := store.Get(batchGasEstimateKey)
	if entity == nil {
		return nil, nil
	}
	estimate := types.MsgEstimateBatchGas{
		Nonce:         nonce,
		TokenContract: tokenContract.GetAddress().Hex(),
		EthSigner:     "",
		Estimate:      0,
	}
	k.cdc.MustUnmarshal(entity, &estimate)
	return &estimate, nil
}

func (k Keeper) SetBatchGasEstimate(ctx context.Context, estimate *types.MsgEstimateBatchGas) ([]byte, error) {
	store := k.GetStore(ctx, types.StoreModulePrefix)
	addr, err := sdk.AccAddressFromBech32(estimate.Metadata.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %s: %w", estimate.Metadata.Creator, err)
	}
	contract, err := types.NewEthAddress(estimate.TokenContract)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %s :%w", estimate.Metadata.Creator, err)
	}
	key, err := types.GetBatchGasEstimateKey(*contract, estimate.Nonce, sdk.ValAddress(addr))
	if err != nil {
		return nil, err
	}
	entity := store.Get(key)
	if entity != nil {
		return nil, fmt.Errorf("estimate already exists: %s", key)
	}
	store.Set(key, k.cdc.MustMarshal(estimate))
	return key, nil
}

func (k Keeper) DeleteBatchGasEstimates(ctx context.Context, batch types.InternalOutgoingTxBatch) error {
	store := k.GetStore(ctx, types.StoreModulePrefix)
	batchGasEstimates, err := k.GetBatchGasEstimateByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract)
	if err != nil {
		return err
	}
	for _, confirm := range batchGasEstimates {
		sender, err := sdk.AccAddressFromBech32(confirm.Metadata.Creator)
		if err != nil {
			return err
		}
		confirmKey, err := types.GetBatchGasEstimateKey(batch.TokenContract, batch.BatchNonce, sdk.ValAddress(sender))
		if err != nil {
			return err
		}
		if store.Has(confirmKey) {
			store.Delete(confirmKey)
		}
	}
	return nil
}

func (k Keeper) IterateBatchGasEstimateByNonceAndTokenContract(ctx context.Context, nonce uint64, tokenContract types.EthAddress, cb func([]byte, types.MsgEstimateBatchGas) bool) error {
	store := k.GetStore(ctx, types.StoreModulePrefix)
	prefix := types.GetBatchGasEstimateNonceContractPrefix(tokenContract, nonce)
	start, end, err := prefixRange(prefix)
	if err != nil {
		return err
	}
	iter := store.Iterator(start, end)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		estimate := types.MsgEstimateBatchGas{
			Nonce:         nonce,
			TokenContract: tokenContract.GetAddress().Hex(),
			EthSigner:     "",
			Estimate:      0,
		}
		k.cdc.MustUnmarshal(iter.Value(), &estimate)
		// cb returns true to stop early
		if cb(iter.Key(), estimate) {
			break
		}
	}
	return nil
}

func (k Keeper) GetBatchGasEstimateByNonceAndTokenContract(ctx context.Context, nonce uint64, tokenContract types.EthAddress) (out []types.MsgEstimateBatchGas, err error) {
	err = k.IterateBatchGasEstimateByNonceAndTokenContract(ctx, nonce, tokenContract, func(_ []byte, msg types.MsgEstimateBatchGas) bool {
		out = append(out, msg)
		return false
	})
	return
}

// TODO: NEEDED?
func (k Keeper) IterateBatchGasEstimates(ctx context.Context, cb func([]byte, types.MsgEstimateBatchGas) (stop bool)) {
	store := k.GetStore(ctx, types.StoreModulePrefix)
	prefixStore := prefix.NewStore(store, types.BatchGasEstimateKey)
	iter := prefixStore.Iterator(nil, nil)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var estimate types.MsgEstimateBatchGas
		k.cdc.MustUnmarshal(iter.Value(), &estimate)

		// cb returns true to stop early
		if cb(iter.Key(), estimate) {
			break
		}
	}
}
