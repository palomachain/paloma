package keeper

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmdb "github.com/cosmos/cosmos-db"
)

func SampleStore(storeKeyName, memStoreKeyName string) (store.CommitMultiStore, *storetypes.KVStoreKey, *storetypes.MemoryStoreKey) {
	storeKey := storetypes.NewKVStoreKey(storeKeyName)
	memStoreKey := storetypes.NewMemoryStoreKey(memStoreKeyName)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.LoadLatestVersion()
	return stateStore, storeKey, memStoreKey
}
