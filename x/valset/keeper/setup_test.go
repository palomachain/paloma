package keeper

import (
	"context"
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
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	params2 "github.com/palomachain/paloma/app/params"
	"github.com/palomachain/paloma/testutil/common"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/palomachain/paloma/x/valset/types/mocks"
	"github.com/stretchr/testify/require"
)

type mockedServices struct {
	StakingKeeper  *mocks.StakingKeeper
	EvmKeeper      *mocks.EvmKeeper
	SlashingKeeper *mocks.SlashingKeeper
}

func newValsetKeeper(t testing.TB) (*Keeper, mockedServices, context.Context) {
	common.SetupPalomaPrefixes()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeKeyService := runtime.NewKVStoreService(storeKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
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
		storeKeyService,
		paramsSubspace,
		ms.StakingKeeper,
		ms.SlashingKeeper,
		"v1.4.0",
		sdk.DefaultPowerReduction,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)

	k.EvmKeeper = ms.EvmKeeper

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil)
	ctx = ctx.WithMultiStore(stateStore).WithGasMeter(storetypes.NewInfiniteGasMeter())

	ctx = ctx.WithLogger(log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
