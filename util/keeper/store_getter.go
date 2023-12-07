package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
)

type StoreGetter interface {
	Store(ctx context.Context) storetypes.KVStore
}

type StoreGetterFn func(ctx context.Context) storetypes.KVStore

func (s StoreGetterFn) Store(ctx context.Context) storetypes.KVStore {
	return s(ctx)
}

func SimpleStoreGetter(s storetypes.KVStore) StoreGetter {
	return StoreGetterFn(func(context.Context) storetypes.KVStore {
		return s
	})
}
