package keeper

import (
	"testing"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	typesparams "cosmossdk.io/x/params/types"
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
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
		storeKey,
		memStoreKey,
		paramsSubspace,
		ms.ValsetKeeper,
		NewRegistry(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
