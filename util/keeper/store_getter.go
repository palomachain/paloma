package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StoreGetter interface {
	Store(ctx sdk.Context) sdk.KVStore
}

type storeGetterFn func(ctx sdk.Context) sdk.KVStore

func (s storeGetterFn) Store(ctx sdk.Context) sdk.KVStore {
	return s(ctx)
}

func SimpleStoreGetter(s sdk.KVStore) StoreGetter {
	return storeGetterFn(func(sdk.Context) sdk.KVStore {
		return s
	})
}
