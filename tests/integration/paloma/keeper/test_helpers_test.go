package keeper_test

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/onsi/ginkgo/v2"
	params2 "github.com/palomachain/paloma/app/params"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	helper "github.com/palomachain/paloma/tests/integration/helper"
	"github.com/palomachain/paloma/x/consensus"
	consensusmodulekeeper "github.com/palomachain/paloma/x/consensus/keeper"
	consensusmoduletypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm"
	evmmodulekeeper "github.com/palomachain/paloma/x/evm/keeper"
	evmmoduletypes "github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/paloma"
	palomakeeper "github.com/palomachain/paloma/x/paloma/keeper"
	palomamoduletypes "github.com/palomachain/paloma/x/paloma/types"
	"github.com/palomachain/paloma/x/scheduler"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	schedulertypes "github.com/palomachain/paloma/x/scheduler/types"
	"github.com/palomachain/paloma/x/valset"
	valsetmodulekeeper "github.com/palomachain/paloma/x/valset/keeper"
	valsetmoduletypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cast"
)

const (
	minimumPigeonVersion = "v1.10.0"
)

type fixture struct {
	ctx               sdk.Context
	codec             codec.Codec
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	consensusKeeper   consensusmodulekeeper.Keeper
	evmKeeper         evmmodulekeeper.Keeper
	valsetKeeper      valsetmodulekeeper.Keeper
	schedulerKeeper   schedulertypes.Keeper
	stakingKeeper     stakingkeeper.Keeper
	paramsKeeper      paramskeeper.Keeper
	palomaKeeper      palomakeeper.Keeper
	upgradeKeeper     upgradekeeper.Keeper
	slashingKeeper    slashingkeeper.Keeper
}

func initFixture(t ginkgo.FullGinkgoTInterface) *fixture {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("paloma", "pub")
	config.SetBech32PrefixForValidator("palomavaloper", "valoperpub")
	// db := dbm.NewMemDB()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey,
		distrtypes.StoreKey, stakingtypes.StoreKey,
		schedulertypes.StoreKey, evmmoduletypes.StoreKey,
		valsetmoduletypes.StoreKey, consensusmoduletypes.StoreKey,
		upgradetypes.StoreKey, slashingtypes.StoreKey,
	)
	encCfg := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		gov.AppModuleBasic{},
		evm.AppModuleBasic{},
		scheduler.AppModuleBasic{},
		valset.AppModuleBasic{},
		consensus.AppModuleBasic{},
		paloma.AppModuleBasic{},
	)

	cdc := encCfg.Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)

	// authority := authtypes.NewModuleAddress(types.ModuleName)

	maccPerms := map[string][]string{
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		schedulertypes.ModuleName:      {authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.Bech32Codec{
			Bech32Prefix: params2.AccountAddressPrefix,
		},
		"paloma",
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appCodec := codec.NewProtoCodec(cdc.InterfaceRegistry())
	legacyAmino := codec.NewLegacyAmino()
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	paramsKeeper := helper.InitParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		helper.BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
		authcodec.NewBech32Codec(params2.ConsNodeAddressPrefix),
	)
	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	valsetKeeper := *valsetmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[valsetmoduletypes.StoreKey]),
		helper.GetSubspace(valsetmoduletypes.ModuleName, paramsKeeper),
		stakingKeeper,
		slashingKeeper,
		minimumPigeonVersion,
		sdk.DefaultPowerReduction,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)
	consensusRegistry := consensusmodulekeeper.NewRegistry()
	consensusKeeper := *consensusmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusmoduletypes.StoreKey]),
		helper.GetSubspace(consensusmoduletypes.ModuleName, paramsKeeper),
		valsetKeeper,
		consensusRegistry,
	)
	evmKeeper := *evmmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evmmoduletypes.StoreKey]),
		consensusKeeper,
		valsetKeeper,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)
	consensusRegistry.Add(
		evmKeeper,
	)
	valsetKeeper.EvmKeeper = evmKeeper
	schedulerKeeper := *keeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[schedulertypes.StoreKey]),
		accountKeeper,
		evmKeeper,
		[]xchain.Bridge{
			evmKeeper,
		},
	)

	appOptions := make(simtestutil.AppOptionsMap, 0)

	appOpts := appOptions
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))

	bApp := *baseapp.NewBaseApp(
		"integration-app",
		logger,
		db.NewMemDB(),
		encCfg.TxConfig.TxDecoder(),
	)

	version.Version = "v5.1.6"
	bApp.SetVersion(version.Version)
	oldVersion := version.Version
	t.Cleanup(func() {
		version.Version = oldVersion
	})

	upgradeKeeper := *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		&bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	semverVersion := bApp.Version()

	if !strings.HasPrefix(semverVersion, "v") {
		semverVersion = fmt.Sprintf("v%s", semverVersion)
	}
	palomaKeeper := palomakeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[palomamoduletypes.StoreKey]),
		helper.GetSubspace(palomamoduletypes.ModuleName, paramsKeeper),
		semverVersion,
		valsetKeeper,
		upgradeKeeper,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)

	palomaKeeper.ExternalChains = []palomamoduletypes.ExternalChainSupporterKeeper{
		evmKeeper,
	}

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	schedulerModule := scheduler.NewAppModule(cdc, schedulerKeeper, accountKeeper, bankKeeper)
	evmModule := evm.NewAppModule(cdc, evmKeeper, accountKeeper, bankKeeper)
	upgradeModule := upgrade.NewAppModule(&upgradeKeeper, address.NewBech32Codec("cosmos"))
	palomaModule := paloma.NewAppModule(cdc, *palomaKeeper, accountKeeper, bankKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:         authModule,
		banktypes.ModuleName:         bankModule,
		schedulertypes.ModuleName:    schedulerModule,
		evmmoduletypes.ModuleName:    evmModule,
		palomamoduletypes.ModuleName: palomaModule,
		upgradetypes.ModuleName:      upgradeModule,
	})

	upgradeKeeper.SetUpgradeHandler(semverVersion, func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (vm module.VersionMap, err error) {
		return
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	queryClient := v1.NewQueryClient(integrationApp.QueryHelper())
	legacyQueryClient := v1beta1.NewQueryClient(integrationApp.QueryHelper())
	return &fixture{
		ctx:               sdkCtx,
		codec:             appCodec,
		queryClient:       queryClient,
		legacyQueryClient: legacyQueryClient,
		consensusKeeper:   consensusKeeper,
		valsetKeeper:      valsetKeeper,
		schedulerKeeper:   schedulerKeeper,
		paramsKeeper:      paramsKeeper,
		evmKeeper:         evmKeeper,
		stakingKeeper:     *stakingKeeper,
		palomaKeeper:      *palomaKeeper,
		upgradeKeeper:     upgradeKeeper,
		slashingKeeper:    slashingKeeper,
	}
}
