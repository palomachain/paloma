package keeper

import (
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmdb "github.com/tendermint/tm-db"
)

func SampleStore(storeKeyName, memStoreKeyName string) (store.CommitMultiStore, *sdk.KVStoreKey, *sdk.MemoryStoreKey) {
	storeKey := sdk.NewKVStoreKey(storeKeyName)
	memStoreKey := storetypes.NewMemoryStoreKey(memStoreKeyName)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	stateStore.LoadLatestVersion()
	return stateStore, storeKey, memStoreKey
}
