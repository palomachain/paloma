package types

import (
	"context"

	cosmosstore "cosmossdk.io/core/store"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

type TreasuryStore interface {
	TreasuryStore(ctx context.Context) storetypes.KVStore
}

type Store struct {
	StoreKey cosmosstore.KVStoreService
}

func (sg Store) TreasuryStore(ctx context.Context) store.KVStore {
	return runtime.KVStoreAdapter(sg.StoreKey.OpenKVStore(ctx))
}
