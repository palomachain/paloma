package keeper_test

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
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
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	params2 "github.com/palomachain/paloma/app/params"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	helper "github.com/palomachain/paloma/tests/integration/helper"
	"github.com/palomachain/paloma/x/consensus"
	consensusmodulekeeper "github.com/palomachain/paloma/x/consensus/keeper"
	consensusmoduletypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm"
	evmmodulekeeper "github.com/palomachain/paloma/x/evm/keeper"
	evmmoduletypes "github.com/palomachain/paloma/x/evm/types"
	gravitymodulekeeper "github.com/palomachain/paloma/x/gravity/keeper"
	gravitymoduletypes "github.com/palomachain/paloma/x/gravity/types"
	"github.com/palomachain/paloma/x/paloma"
	palomamoduletypes "github.com/palomachain/paloma/x/paloma/types"
	"github.com/palomachain/paloma/x/scheduler"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	schedulertypes "github.com/palomachain/paloma/x/scheduler/types"
	treasurymoduletypes "github.com/palomachain/paloma/x/treasury/types"
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
}

func init() {
	params2.SetAddressConfig()
}

func initFixture(t ginkgo.FullGinkgoTInterface) *fixture {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		ibcexported.StoreKey,
		ibctransfertypes.StoreKey,
		ibcfeetypes.StoreKey,
		icahosttypes.StoreKey,
		icacontrollertypes.StoreKey,
		capabilitytypes.StoreKey,
		consensusmoduletypes.StoreKey,
		valsetmoduletypes.StoreKey,
		treasurymoduletypes.StoreKey,
		evmmoduletypes.StoreKey,
		gravitymoduletypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		gov.AppModuleBasic{},
		evm.AppModuleBasic{},
		staking.AppModuleBasic{},
		scheduler.AppModuleBasic{},
		valset.AppModuleBasic{},
		consensus.AppModuleBasic{},
	).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)

	maccPerms := map[string][]string{
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		schedulertypes.ModuleName:      {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		ibcfeetypes.ModuleName:         nil,
		icatypes.ModuleName:            nil,
	}

	memKeys := storetypes.NewMemoryStoreKeys(
		capabilitytypes.MemStoreKey,
		valsetmoduletypes.MemStoreKey,
		consensusmoduletypes.MemStoreKey,
		evmmoduletypes.MemStoreKey,
		treasurymoduletypes.MemStoreKey,
		palomamoduletypes.MemStoreKey,
	)

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
	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	skipUpgradeHeights := map[int64]bool{}
	appOpts := db.OptionsMap{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
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
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	bApp := *baseapp.NewBaseApp(
		"integration-app",
		logger,
		db.NewMemDB(),
		encCfg.TxConfig.TxDecoder(),
	)

	version.Version = "v5.1.6"
	bApp.SetVersion(version.Version)

	upgradeKeeper := *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		&bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	IBCKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		helper.GetSubspace(ibcexported.ModuleName, paramsKeeper),
		stakingKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	IBCFeeKeeper := ibcfeekeeper.NewKeeper(
		appCodec,
		keys[ibcfeetypes.StoreKey],
		IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		IBCKeeper.ChannelKeeper,
		IBCKeeper.PortKeeper,
		accountKeeper,
		bankKeeper,
	)
	scopedTransferKeeper := capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		helper.GetSubspace(ibctransfertypes.ModuleName, paramsKeeper),
		IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		IBCKeeper.ChannelKeeper,
		IBCKeeper.PortKeeper,
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
	gravityKeeper := gravitymodulekeeper.NewKeeper(
		appCodec,
		accountKeeper,
		stakingKeeper,
		bankKeeper,
		slashingKeeper,
		distrKeeper,
		transferKeeper,
		evmKeeper,
		gravitymodulekeeper.NewGravityStoreGetter(keys[gravitymoduletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(params2.ValidatorAddressPrefix),
	)
	evmKeeper.Gravity = gravityKeeper
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

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	schedulerModule := scheduler.NewAppModule(cdc, schedulerKeeper, accountKeeper, bankKeeper)
	evmModule := evm.NewAppModule(cdc, evmKeeper, accountKeeper, bankKeeper)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, helper.GetSubspace(stakingtypes.ModuleName, paramsKeeper))
	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:      authModule,
		banktypes.ModuleName:      bankModule,
		schedulertypes.ModuleName: schedulerModule,
		evmmoduletypes.ModuleName: evmModule,
		stakingtypes.ModuleName:   stakingModule,
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
	}
}
