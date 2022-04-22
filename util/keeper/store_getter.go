package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StoreGetter interface {
	Store(ctx sdk.Context) sdk.KVStore
}

type StoreGetterFn func(ctx sdk.Context) sdk.KVStore

func (s StoreGetterFn) Store(ctx sdk.Context) sdk.KVStore {
	return s(ctx)
}

func SimpleStoreGetter(s sdk.KVStore) StoreGetter {
	return StoreGetterFn(func(sdk.Context) sdk.KVStore {
		return s
	})
}
