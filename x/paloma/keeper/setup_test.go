package keeper_test

import (
	"time"

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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/v2/app/params"
	"github.com/palomachain/paloma/v2/testutil"
	"github.com/palomachain/paloma/v2/x/paloma/keeper"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"github.com/palomachain/paloma/v2/x/paloma/types/mocks"
	"github.com/stretchr/testify/require"
)

const testBondDenom = "testgrain"

type mockedServices struct {
	AccountKeeper  *mocks.AccountKeeper
	BankKeeper     *mocks.BankKeeper
	FeegrantKeeper *mocks.FeegrantKeeper
	ValsetKeeper   *mocks.ValsetKeeper
	UpgradeKeeper  *mocks.UpgradeKeeper
}

func newMockedKeeper(t testutil.TB) (*keeper.Keeper, mockedServices, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)

	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(registry)

	types.RegisterInterfaces(registry)

	ms := mockedServices{
		AccountKeeper:  mocks.NewAccountKeeper(t),
		BankKeeper:     mocks.NewBankKeeper(t),
		FeegrantKeeper: mocks.NewFeegrantKeeper(t),
		ValsetKeeper:   mocks.NewValsetKeeper(t),
		UpgradeKeeper:  mocks.NewUpgradeKeeper(t),
	}

	paramsSubspace := paramstypes.NewSubspace(appCodec,
		types.Amino,
		storeKey,
		memStoreKey,
		"PalomaParams",
	)

	k := keeper.NewKeeper(
		appCodec,
		storeService,
		paramsSubspace,
		"0.0.0",
		testBondDenom,
		ms.AccountKeeper,
		ms.BankKeeper,
		ms.FeegrantKeeper,
		ms.ValsetKeeper,
		ms.UpgradeKeeper,
		authcodec.NewBech32Codec(params.ValidatorAddressPrefix),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	header := tmproto.Header{
		Time: time.Now(),
	}
	ctx := sdk.NewContext(stateStore, header, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
