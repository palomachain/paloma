package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	testtypes "github.com/volumefi/cronchain/x/scheduler/testdata/types"
	"github.com/volumefi/cronchain/x/scheduler/types"
)

type storeGetterFn func(ctx sdk.Context) sdk.KVStore

func (s storeGetterFn) Store(ctx sdk.Context) sdk.KVStore {
	return s(ctx)
}

func simpleStoreGetter(s sdk.KVStore) storeGetter {
	return storeGetterFn(func(sdk.Context) sdk.KVStore {
		return s
	})
}

func TestConcensusQueueMessage(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	stateStore.LoadLatestVersion()

	registry := types.ModuleCdc.InterfaceRegistry()
	registry.RegisterInterface(
		"volumefi.tests.SimpleMessage",
		(*sdk.Msg)(nil),
		&testtypes.SimpleMessage{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil), &testtypes.SimpleMessage{})
	types.RegisterInterfaces(registry)

	sg := simpleStoreGetter(stateStore.GetKVStore(storeKey))
	cq := concensusQueue[*testtypes.SimpleMessage]{
		storeGetter: sg,
		ider: ider{
			sg:    sg,
			idKey: []byte("simple-message"),
		},
		cdc: types.ModuleCdc,
	}
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil)

	msg := &testtypes.SimpleMessage{
		Sender: "bob",
		Hello:  "HEY",
		World:  "WORLD",
	}

	err := cq.Put(ctx, msg)

	assert.NoError(t, err)
	cq.GetMsgsForSigning(ctx, sdk.ValAddress(`bob`))
}
