package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StoreGetter interface {
	Store(ctx sdk.Context) storetypes.KVStore
}

type StoreGetterFn func(ctx sdk.Context) storetypes.KVStore

func (s StoreGetterFn) Store(ctx sdk.Context) storetypes.KVStore {
	return s(ctx)
}

func SimpleStoreGetter(s storetypes.KVStore) StoreGetter {
	return StoreGetterFn(func(sdk.Context) storetypes.KVStore {
		return s
	})
}
