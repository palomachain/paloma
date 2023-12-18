package keeper

import (
	"os"
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/palomachain/paloma/x/valset/types/mocks"
	"github.com/stretchr/testify/require"
)

type mockedServices struct {
	StakingKeeper  *mocks.StakingKeeper
	EvmKeeper      *mocks.EvmKeeper
	SlashingKeeper *mocks.SlashingKeeper
}

func newValsetKeeper(t testing.TB) (*Keeper, mockedServices, sdk.Context) {
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

	paramsSubspace := typesparams.NewSubspace(appCodec,
		types.Amino,
		storeKey,
		memStoreKey,
		"ValsetParams",
	)

	ms := mockedServices{
		StakingKeeper:  mocks.NewStakingKeeper(t),
		EvmKeeper:      mocks.NewEvmKeeper(t),
		SlashingKeeper: mocks.NewSlashingKeeper(t),
	}
	k := NewKeeper(
		appCodec,
		storeKey,
		memStoreKey,
		paramsSubspace,
		ms.StakingKeeper,
		ms.SlashingKeeper,
		"v1.4.0",
		sdk.DefaultPowerReduction,
	)

	k.EvmKeeper = ms.EvmKeeper

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil)
	ctx = ctx.WithMultiStore(stateStore).WithGasMeter(sdk.NewInfiniteGasMeter())

	ctx = ctx.WithLogger(log.NewTMJSONLogger(os.Stdout))

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
