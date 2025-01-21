package helper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
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
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
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
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/onsi/ginkgo/v2"
	"github.com/palomachain/paloma/v2/app"
	params2 "github.com/palomachain/paloma/v2/app/params"
	xchain "github.com/palomachain/paloma/v2/internal/x-chain"
	"github.com/palomachain/paloma/v2/x/consensus"
	consensusmodulekeeper "github.com/palomachain/paloma/v2/x/consensus/keeper"
	consensusmoduletypes "github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/palomachain/paloma/v2/x/evm"
	evmmodulekeeper "github.com/palomachain/paloma/v2/x/evm/keeper"
	evmmoduletypes "github.com/palomachain/paloma/v2/x/evm/types"
	"github.com/palomachain/paloma/v2/x/metrix"
	metrixmodulekeeper "github.com/palomachain/paloma/v2/x/metrix/keeper"
	metrixmoduletypes "github.com/palomachain/paloma/v2/x/metrix/types"
	"github.com/palomachain/paloma/v2/x/paloma"
	palomakeeper "github.com/palomachain/paloma/v2/x/paloma/keeper"
	palomamoduletypes "github.com/palomachain/paloma/v2/x/paloma/types"
	"github.com/palomachain/paloma/v2/x/scheduler"
	"github.com/palomachain/paloma/v2/x/scheduler/keeper"
	schedulertypes "github.com/palomachain/paloma/v2/x/scheduler/types"
	skywaymodulekeeper "github.com/palomachain/paloma/v2/x/skyway/keeper"
	skywaymoduletypes "github.com/palomachain/paloma/v2/x/skyway/types"
	treasurykeeper "github.com/palomachain/paloma/v2/x/treasury/keeper"
	treasurymoduletypes "github.com/palomachain/paloma/v2/x/treasury/types"
	"github.com/palomachain/paloma/v2/x/valset"
	valsetmodulekeeper "github.com/palomachain/paloma/v2/x/valset/keeper"
	valsetmoduletypes "github.com/palomachain/paloma/v2/x/valset/types"
	"github.com/spf13/cast"
)

type Fixture struct {
	Ctx   sdk.Context
	Codec codec.Codec

	QueryClient       v1.QueryClient
	LegacyQueryClient v1beta1.QueryClient

	ConsensusKeeper consensusmodulekeeper.Keeper
	EvmKeeper       evmmodulekeeper.Keeper
	MetrixKeeper    metrixmodulekeeper.Keeper
	PalomaKeeper    palomakeeper.Keeper
	ParamsKeeper    paramskeeper.Keeper
	SchedulerKeeper keeper.Keeper
	SlashingKeeper  slashingkeeper.Keeper
	StakingKeeper   stakingkeeper.Keeper
	TreasuryKeeper  treasurykeeper.Keeper
	UpgradeKeeper   upgradekeeper.Keeper
	ValsetKeeper    valsetmodulekeeper.Keeper
}

func InitFixture(t ginkgo.FullGinkgoTInterface) *Fixture {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("paloma", "pub")
	config.SetBech32PrefixForValidator("palomavaloper", "valoperpub")

	// ****************************************************************
	// * Store keys (IAVL, trancient, mem)
	// ****************************************************************
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		capabilitytypes.StoreKey,
		consensusmoduletypes.StoreKey,
		distrtypes.StoreKey,
		evidencetypes.StoreKey,
		evmmoduletypes.StoreKey,
		feegrant.StoreKey,
		govtypes.StoreKey,
		ibcexported.StoreKey,
		ibcfeetypes.StoreKey,
		ibctransfertypes.StoreKey,
		icacontrollertypes.StoreKey,
		icahosttypes.StoreKey,
		metrixmoduletypes.StoreKey,
		minttypes.StoreKey,
		palomamoduletypes.StoreKey,
		paramstypes.StoreKey,
		schedulertypes.StoreKey,
		skywaymoduletypes.StoreKey,
		slashingtypes.StoreKey,
		stakingtypes.StoreKey,
		treasurymoduletypes.StoreKey,
		upgradetypes.StoreKey,
		valsetmoduletypes.StoreKey,
	)
	memKeys := storetypes.NewMemoryStoreKeys(
		capabilitytypes.MemStoreKey,
		consensusmoduletypes.MemStoreKey,
		evmmoduletypes.MemStoreKey,
		metrixmoduletypes.MemStoreKey,
		palomamoduletypes.MemStoreKey,
		treasurymoduletypes.MemStoreKey,
		valsetmoduletypes.MemStoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		consensus.AppModuleBasic{},
		evm.AppModuleBasic{},
		gov.AppModuleBasic{},
		metrix.AppModuleBasic{},
		paloma.AppModuleBasic{},
		scheduler.AppModuleBasic{},
		staking.AppModuleBasic{},
		valset.AppModuleBasic{},
	)

	// ****************************************************************
	// * Storage mounting and logging
	// ****************************************************************
	cdc := encCfg.Codec
	logger := log.NewNopLogger()
	cms := integration.CreateMultiStore(keys, logger)
	for _, key := range memKeys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeMemory, nil)
	}
	for _, key := range tkeys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeTransient, nil)
	}
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	appCodec := codec.NewProtoCodec(cdc.InterfaceRegistry())
	legacyAmino := codec.NewLegacyAmino()
	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)
	maccPerms := map[string][]string{
		distrtypes.ModuleName:          nil,
		ibcfeetypes.ModuleName:         nil,
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		icatypes.ModuleName:            nil,
		minttypes.ModuleName:           {authtypes.Minter},
		schedulertypes.ModuleName:      {authtypes.Burner},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	// ****************************************************************
	// * Base app
	// ****************************************************************
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
	semverVersion := bApp.Version()
	if !strings.HasPrefix(semverVersion, "v") {
		semverVersion = fmt.Sprintf("v%s", semverVersion)
	}

	// ****************************************************************
	// * Base modules
	// ****************************************************************
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

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)

	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	paramsKeeper := initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	feegrantKeeper := feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), accountKeeper)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedTransferKeeper := capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

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

	upgradeKeeper := *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		&bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		getSubspace(ibcexported.ModuleName, paramsKeeper),
		stakingKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	ibcFeeKeeper := ibcfeekeeper.NewKeeper(
		appCodec,
		keys[ibcfeetypes.StoreKey],
		ibcKeeper.ChannelKeeper, // may be replaced with IBC middleware
		ibcKeeper.ChannelKeeper,
		ibcKeeper.PortKeeper,
		accountKeeper,
		bankKeeper,
	)

	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		getSubspace(ibctransfertypes.ModuleName, paramsKeeper),
		ibcFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		ibcKeeper.ChannelKeeper,
		ibcKeeper.PortKeeper,
		accountKeeper,
		bankKeeper,
		scopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	distrKeeper := distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// ****************************************************************
	// * Custom modules
	// ****************************************************************
	var tk treasurykeeper.Keeper = treasurykeeper.Keeper{}
	var evmKeeper *evmmodulekeeper.Keeper = &evmmodulekeeper.Keeper{}
	consensusRegistry := consensusmodulekeeper.NewRegistry()

	metrixKeeper := metrixmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[metrixmoduletypes.StoreKey]),
		getSubspace(metrixmoduletypes.ModuleName, paramsKeeper),
		slashingKeeper,
		stakingKeeper,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)

	tk = *treasurykeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[treasurymoduletypes.StoreKey]),
		getSubspace(treasurymoduletypes.ModuleName, paramsKeeper),
		bankKeeper,
		accountKeeper,
		evmKeeper,
	)

	valsetKeeper := *valsetmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[valsetmoduletypes.StoreKey]),
		getSubspace(valsetmoduletypes.ModuleName, paramsKeeper),
		stakingKeeper,
		slashingKeeper,
		sdk.DefaultPowerReduction,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)

	consensusKeeper := *consensusmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusmoduletypes.StoreKey]),
		getSubspace(consensusmoduletypes.ModuleName, paramsKeeper),
		valsetKeeper,
		consensusRegistry,
		&tk,
	)

	*evmKeeper = *evmmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evmmoduletypes.StoreKey]),
		consensusKeeper,
		valsetKeeper,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
		metrixKeeper,
		tk,
	)

	schedulerKeeper := *keeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[schedulertypes.StoreKey]),
		accountKeeper,
		evmKeeper,
		[]xchain.Bridge{
			evmKeeper,
		},
	)

	skywayKeeper := skywaymodulekeeper.NewKeeper(
		appCodec,
		accountKeeper,
		stakingKeeper,
		bankKeeper,
		slashingKeeper,
		distrKeeper,
		transferKeeper,
		evmKeeper,
		consensusKeeper,
		nil,
		nil,
		skywaymodulekeeper.NewSkywayStoreGetter(keys[skywaymoduletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)

	palomaKeeper := palomakeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[palomamoduletypes.StoreKey]),
		getSubspace(palomamoduletypes.ModuleName, paramsKeeper),
		semverVersion,
		app.BondDenom,
		accountKeeper,
		bankKeeper,
		feegrantKeeper,
		valsetKeeper,
		upgradeKeeper,
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// ****************************************************************
	// * Late DI
	// ****************************************************************
	evmKeeper.Skyway = skywayKeeper
	consensusRegistry.Add(
		evmKeeper,
	)
	valsetKeeper.EvmKeeper = evmKeeper
	valsetKeeper.SnapshotListeners = []valsetmoduletypes.OnSnapshotBuiltListener{evmKeeper, &metrixKeeper}

	palomaKeeper.ExternalChains = []palomamoduletypes.ExternalChainSupporterKeeper{
		evmKeeper,
	}

	// ****************************************************************
	// * Module building
	// ****************************************************************
	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	schedulerModule := scheduler.NewAppModule(cdc, schedulerKeeper, accountKeeper, bankKeeper)
	metrixModule := metrix.NewAppModule(cdc, metrixKeeper)
	evmModule := evm.NewAppModule(cdc, *evmKeeper, accountKeeper, bankKeeper)
	upgradeModule := upgrade.NewAppModule(&upgradeKeeper, address.NewBech32Codec("cosmos"))
	palomaModule := paloma.NewAppModule(cdc, *palomaKeeper, accountKeeper, bankKeeper)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, getSubspace(stakingtypes.ModuleName, paramsKeeper))

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:         authModule,
		banktypes.ModuleName:         bankModule,
		schedulertypes.ModuleName:    schedulerModule,
		evmmoduletypes.ModuleName:    evmModule,
		palomamoduletypes.ModuleName: palomaModule,
		upgradetypes.ModuleName:      upgradeModule,
		metrixmoduletypes.ModuleName: metrixModule,
		stakingtypes.ModuleName:      stakingModule,
	})

	upgradeKeeper.SetUpgradeHandler(semverVersion, func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (vm module.VersionMap, err error) {
		return
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	queryClient := v1.NewQueryClient(integrationApp.QueryHelper())
	legacyQueryClient := v1beta1.NewQueryClient(integrationApp.QueryHelper())
	metrixKeeper.UpdateUptime(integrationApp.Context())
	return &Fixture{
		Ctx:               sdkCtx,
		Codec:             appCodec,
		QueryClient:       queryClient,
		LegacyQueryClient: legacyQueryClient,
		ConsensusKeeper:   consensusKeeper,
		ValsetKeeper:      valsetKeeper,
		SchedulerKeeper:   schedulerKeeper,
		ParamsKeeper:      paramsKeeper,
		EvmKeeper:         *evmKeeper,
		StakingKeeper:     *stakingKeeper,
		PalomaKeeper:      *palomaKeeper,
		UpgradeKeeper:     upgradeKeeper,
		SlashingKeeper:    slashingKeeper,
		MetrixKeeper:      metrixKeeper,
		TreasuryKeeper:    tk,
	}
}
