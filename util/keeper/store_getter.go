package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StoreGetter interface {
	Store(ctx context.Context) sdk.KVStore
}

type StoreGetterFn func(ctx context.Context) sdk.KVStore

func (s StoreGetterFn) Store(ctx context.Context) sdk.KVStore {
	return s(ctx)
}

func SimpleStoreGetter(s sdk.KVStore) StoreGetter {
	return StoreGetterFn(func(context.Context) sdk.KVStore {
		return s
	})
}
