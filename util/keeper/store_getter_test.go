package keeper

import (
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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
