package keeper

import (
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmdb "github.com/cometbft/cometbft-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func SampleStore(storeKeyName, memStoreKeyName string) (store.CommitMultiStore, *storetypes.KVStoreKey, *storetypes.MemoryStoreKey) {
	storeKey := sdk.NewKVStoreKey(storeKeyName)
	memStoreKey := storetypes.NewMemoryStoreKey(memStoreKeyName)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.LoadLatestVersion()
	return stateStore, storeKey, memStoreKey
}
