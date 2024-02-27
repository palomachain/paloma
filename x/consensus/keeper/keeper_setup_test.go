package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/consensus/types/mocks"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
	"github.com/stretchr/testify/require"
)

type mockedServices struct {
	ValsetKeeper *mocks.ValsetKeeper
}

func newConsensusKeeper(t testing.TB) (*Keeper, mockedServices, sdk.Context) {
	logger := log.NewNopLogger()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(registry)

	types.RegisterInterfaces(registry)
	evmtypes.RegisterInterfaces(registry)

	registry.RegisterImplementations((*types.ConsensusMsg)(nil),
		&types.SimpleMessage{},
	)

	paramsSubspace := typesparams.NewSubspace(appCodec,
		types.Amino,
		storeKey,
		memStoreKey,
		"ConsensusParams",
	)
	ms := mockedServices{
		ValsetKeeper: mocks.NewValsetKeeper(t),
	}
	k := NewKeeper(
		appCodec,
		runtime.NewKVStoreService(storeKey),
		paramsSubspace,
		ms.ValsetKeeper,
		NewRegistry(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
