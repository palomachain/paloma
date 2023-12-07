package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// "github.com/CosmWasm/wasmd/x/wasm"
	// wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"cosmossdk.io/log"
	// "github.com/CosmWasm/wasmd/x/wasm"
	// "github.com/CosmWasm/wasmd/x/wasm"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	// xchain "github.com/palomachain/paloma/internal/x-chain"
	// "github.com/palomachain/paloma/x/evm"

	"github.com/cosmos/cosmos-sdk/std"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/ibc-go/modules/capability"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"

	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	palomamempool "github.com/palomachain/paloma/app/mempool"
	appparams "github.com/palomachain/paloma/app/params"

	// xchain "github.com/palomachain/paloma/internal/x-chain"
	consensusmodule "github.com/palomachain/paloma/x/consensus"
	consensusmodulekeeper "github.com/palomachain/paloma/x/consensus/keeper"
	consensusmoduletypes "github.com/palomachain/paloma/x/consensus/types"

	// "github.com/palomachain/paloma/x/evm"
	// evmclient "github.com/palomachain/paloma/x/evm/client"
	// evmmodulekeeper "github.com/palomachain/paloma/x/evm/keeper"
	// evmmoduletypes "github.com/palomachain/paloma/x/evm/types"
	// gravitymodule "github.com/palomachain/paloma/x/gravity"
	// gravityclient "github.com/palomachain/paloma/x/gravity/client"
	// gravitymodulekeeper "github.com/palomachain/paloma/x/gravity/keeper"
	// gravitymoduletypes "github.com/palomachain/paloma/x/gravity/types"
	// palomamodule "github.com/palomachain/paloma/x/paloma"
	// palomamodulekeeper "github.com/palomachain/paloma/x/paloma/keeper"
	// palomamoduletypes "github.com/palomachain/paloma/x/paloma/types"
	// schedulermodule "github.com/palomachain/paloma/x/scheduler"
	// schedulermodulekeeper "github.com/palomachain/paloma/x/scheduler/keeper"
	// schedulermoduletypes "github.com/palomachain/paloma/x/scheduler/types"
	// treasurymodule "github.com/palomachain/paloma/x/treasury"
	// treasuryclient "github.com/palomachain/paloma/x/treasury/client"
	// treasurymodulekeeper "github.com/palomachain/paloma/x/treasury/keeper"
	// treasurymoduletypes "github.com/palomachain/paloma/x/treasury/types"
	// valsetmodule "github.com/palomachain/paloma/x/valset"
	// valsetmodulekeeper "github.com/palomachain/paloma/x/valset/keeper"
	// valsetmoduletypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cast"
)

const (
	Name = "paloma"

	wasmAvailableCapabilities = "iterator,staking,stargate,paloma"

	minimumPigeonVersion = "v1.10.0"
)

func getGovProposalHandlers() []govclient.ProposalHandler {
	return []govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		// evmclient.ProposalHandler,
		// gravityclient.ProposalHandler,
		// treasuryclient.ProposalHandler,
	}
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		// gravitymoduletypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		ibcfeetypes.ModuleName:      nil,
		icatypes.ModuleName:         nil,
		// wasm.ModuleName:                {authtypes.Burner},
		// treasurymoduletypes.ModuleName: {authtypes.Burner, authtypes.Minter},
	}
)

var _ servertypes.Application = (*App)(nil)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	invCheckPeriod    uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper

	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedIBCFeeKeeper        capabilitykeeper.ScopedKeeper
	// ScopedWasmKeeper          capabilitykeeper.ScopedKeeper
	ScopedConsensusKeeper capabilitykeeper.ScopedKeeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper          ibcfeekeeper.Keeper
	ICAControllerKeeper   icacontrollerkeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper

	// SchedulerKeeper schedulermodulekeeper.Keeper
	ConsensusKeeper consensusmodulekeeper.Keeper
	// ValsetKeeper    valsetmodulekeeper.Keeper
	// PalomaKeeper    palomamodulekeeper.Keeper
	// TreasuryKeeper  treasurymodulekeeper.Keeper
	// EvmKeeper       evmmodulekeeper.Keeper
	// GravityKeeper   gravitymodulekeeper.Keeper
	// wasmKeeper      wasm.Keeper

	// ModuleManager is the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// sm is the simulation manager
	sm *module.SimulationManager

	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	encodingCfg appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)
	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetTxEncoder(txConfig.TxEncoder())
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	nonceMempool := palomamempool.DefaultPriorityMempool()
	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, bApp)

	bApp.SetMempool(nonceMempool)
	bApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	bApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())

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
		// schedulermoduletypes.StoreKey,
		consensusmoduletypes.StoreKey,
		// valsetmoduletypes.StoreKey,
		// treasurymoduletypes.StoreKey,
		// // wasm.StoreKey,
		// evmmoduletypes.StoreKey,
		// gravitymoduletypes.StoreKey,
		consensusparamtypes.StoreKey,
		crisistypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(
		capabilitytypes.MemStoreKey,
		// valsetmoduletypes.MemStoreKey,
		consensusmoduletypes.MemStoreKey,
		// evmmoduletypes.MemStoreKey,
		// schedulermoduletypes.MemStoreKey,
		// treasurymoduletypes.MemStoreKey,
		// palomamoduletypes.MemStoreKey,
	)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	// scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	scopedConsensusKeeper := app.CapabilityKeeper.ScopeToModule(consensusmoduletypes.ModuleName)

	app.CapabilityKeeper.Seal()

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.StakingKeeper.ValidatorAddressCodec(),
		app.StakingKeeper.ConsensusAddressCodec(),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper, app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.invCheckPeriod = cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		app.invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	// set the governance module account as the authority for conducting upgrades
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec,
		keys[ibcfeetypes.StoreKey],
		app.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	semverVersion := app.Version()
	if !strings.HasPrefix(semverVersion, "v") {
		semverVersion = fmt.Sprintf("v%s", semverVersion)
	}

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	// app.ValsetKeeper = *valsetmodulekeeper.NewKeeper(
	// 	appCodec,
	// 	keys[valsetmoduletypes.StoreKey],
	// 	memKeys[valsetmoduletypes.MemStoreKey],
	// 	app.GetSubspace(valsetmoduletypes.ModuleName),
	// 	app.StakingKeeper,
	// 	minimumPigeonVersion,
	// 	sdk.DefaultPowerReduction,
	// )

	consensusRegistry := consensusmodulekeeper.NewRegistry()

	app.ConsensusKeeper = *consensusmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusmoduletypes.StoreKey]),
		memKeys[consensusmoduletypes.MemStoreKey],
		app.GetSubspace(consensusmoduletypes.ModuleName),
		nil, //consensusRegistry,
		nil, //authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// app.EvmKeeper = *evmmodulekeeper.NewKeeper(
	// 	appCodec,
	// 	keys[evmmoduletypes.StoreKey],
	// 	memKeys[evmmoduletypes.MemStoreKey],
	// 	app.GetSubspace(evmmoduletypes.ModuleName),
	// app.ConsensusKeeper,
	// 	app.ValsetKeeper,
	// )
	// app.ValsetKeeper.SnapshotListeners = []valsetmoduletypes.OnSnapshotBuiltListener{
	// 	app.EvmKeeper,
	// }
	// app.ValsetKeeper.EvmKeeper = app.EvmKeeper

	// app.GravityKeeper = gravitymodulekeeper.NewKeeper(
	// 	appCodec,
	// 	app.GetSubspace(gravitymoduletypes.ModuleName),
	// 	app.AccountKeeper,
	// 	app.StakingKeeper,
	// 	app.BankKeeper,
	// 	app.SlashingKeeper,
	// 	app.DistrKeeper,
	// 	app.EvmKeeper,
	// 	app.TransferKeeper,
	// 	gravitymodulekeeper.NewGravityStoreGetter(keys[gravitymoduletypes.StoreKey]),
	// )

	// app.PalomaKeeper = *palomamodulekeeper.NewKeeper(
	// 	appCodec,
	// 	keys[palomamoduletypes.StoreKey],
	// 	memKeys[palomamoduletypes.MemStoreKey],
	// 	app.GetSubspace(palomamoduletypes.ModuleName),
	// 	semverVersion,
	// 	app.ValsetKeeper,
	// 	app.PalomaKeeper.Upgrade,
	// )

	// app.PalomaKeeper.ExternalChains = []palomamoduletypes.ExternalChainSupporterKeeper{
	// 	app.EvmKeeper,
	// }

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		nil,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// app.SchedulerKeeper = *schedulermodulekeeper.NewKeeper(
	// 	appCodec,
	// 	keys[schedulermoduletypes.StoreKey],
	// 	memKeys[schedulermoduletypes.MemStoreKey],
	// 	app.GetSubspace(schedulermoduletypes.ModuleName),
	// 	app.AccountKeeper,
	// 	app.EvmKeeper,
	// 	[]xchain.Bridge{
	// 		app.EvmKeeper,
	// 	},
	// )

	// app.TreasuryKeeper = *treasurymodulekeeper.NewKeeper(
	// 	appCodec,
	// 	keys[treasurymoduletypes.StoreKey],
	// 	memKeys[treasurymoduletypes.MemStoreKey],
	// 	app.GetSubspace(schedulermoduletypes.ModuleName),
	// 	app.BankKeeper,
	// 	app.AccountKeeper,
	// 	app.SchedulerKeeper,
	// )

	app.ScopedConsensusKeeper = scopedConsensusKeeper

	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))
	// AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))
	// AddRoute(evmmoduletypes.RouterKey, evm.NewReferenceChainReferenceIDProposalHandler(app.EvmKeeper)).
	// AddRoute(gravitymoduletypes.RouterKey, gravitymodulekeeper.NewGravityProposalHandler(app.GravityKeeper)).
	// AddRoute(treasurymoduletypes.RouterKey, treasurymodule.NewFeeProposalHandler(app.TreasuryKeeper))

	// Example of setting gov params:
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govtypes.Config{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	// wasmDir := filepath.Join(homePath, "data")
	// wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	// if err != nil {
	// 	panic("error while reading wasm config: " + err.Error())
	// }

	// app.wasmKeeper = wasm.NewKeeper(
	// 	appCodec,
	// 	keys[wasm.StoreKey],
	// 	app.AccountKeeper,
	// 	app.BankKeeper,
	// 	app.StakingKeeper,
	// 	distrkeeper.NewQuerier(app.DistrKeeper),
	// 	// app.IBCKeeper.ChannelKeeper,
	// 	// app.IBCKeeper.ChannelKeeper,
	// 	// &app.IBCKeeper.PortKeeper,
	// 	scopedWasmKeeper,
	// 	// app.TransferKeeper,
	// 	app.MsgServiceRouter(),
	// 	app.GRPCQueryRouter(),
	// 	wasmDir,
	// 	wasmConfig,
	// 	wasmAvailableCapabilities,
	// 	authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	// 	wasmkeeper.WithMessageHandlerDecorator(func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
	// 		return wasmkeeper.NewMessageHandlerChain(
	// 			old,
	// 			app.SchedulerKeeper.ExecuteWasmJobEventListener(),
	// 		)
	// 	}),
	// )

	// register wasm gov proposal types
	// enabledProposals := GetEnabledProposals()
	// if len(enabledProposals) != 0 {
	// 	govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.wasmKeeper, enabledProposals))
	// }

	// Create Transfer Stack
	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, app.IBCFeeKeeper)

	// Create Interchain Accounts Stack
	// SendPacket, since it is originating from the application to core IBC:
	// icaAuthModuleKeeper.SendTx -> icaController.SendPacket -> fee.SendPacket -> channel.SendPacket
	var icaControllerStack porttypes.IBCModule

	// integration point for custom authentication modules
	// see https://medium.com/the-interchain-foundation/ibc-go-v6-changes-to-interchain-accounts-and-how-it-impacts-your-chain-806c185300d7
	var noAuthzModule porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddleware(noAuthzModule, app.ICAControllerKeeper)
	icaControllerStack = ibcfee.NewIBCMiddleware(icaControllerStack, app.IBCFeeKeeper)

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(app.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, app.IBCFeeKeeper)

	// Create fee enabled wasm ibc Stack
	// var wasmStack porttypes.IBCModule
	// wasmStack = wasm.NewIBCHandler(app.wasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCFeeKeeper)
	// wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.IBCFeeKeeper)

	// Create static IBC router, add app routes, then set and seal it
	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		// AddRoute(wasm.ModuleName, wasmStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// app.ScopedWasmKeeper = scopedWasmKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper

	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	// evmModule := evm.NewAppModule(appCodec, app.EvmKeeper, app.AccountKeeper, app.BankKeeper)
	consensusModule := consensusmodule.NewAppModule(appCodec, app.ConsensusKeeper, app.AccountKeeper, app.BankKeeper)
	// valsetModule := valsetmodule.NewAppModule(appCodec, app.ValsetKeeper, app.AccountKeeper, app.BankKeeper)
	// schedulerModule := schedulermodule.NewAppModule(appCodec, app.SchedulerKeeper, app.AccountKeeper, app.BankKeeper)
	// palomaModule := palomamodule.NewAppModule(appCodec, app.PalomaKeeper, app.AccountKeeper, app.BankKeeper)
	// gravityModule := gravitymodule.NewAppModule(appCodec, app.GravityKeeper, app.BankKeeper)
	// treasuryModule := treasurymodule.NewAppModule(appCodec, app.TreasuryKeeper, app.AccountKeeper, app.BankKeeper)
	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		// schedulerModule,
		consensusModule,
		// valsetModule,
		// evmModule,
		// gravityModule,
		// palomaModule,
		// treasuryModule,
		// wasm.NewAppModule(appCodec, &app.wasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasm.ModuleName)),
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		ibcfee.NewAppModule(app.IBCFeeKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		feegrant.ModuleName,
		// schedulermoduletypes.ModuleName,
		consensusmoduletypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		authtypes.ModuleName,
		vestingtypes.ModuleName,
		genutiltypes.ModuleName,
		// valsetmoduletypes.ModuleName,
		// palomamoduletypes.ModuleName,
		// wasm.ModuleName,
		// evmmoduletypes.ModuleName,
		// gravitymoduletypes.ModuleName,
		// treasurymoduletypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		// schedulermoduletypes.ModuleName,
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		consensusmoduletypes.ModuleName,
		crisistypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		authtypes.ModuleName,
		vestingtypes.ModuleName,
		genutiltypes.ModuleName,
		// valsetmoduletypes.ModuleName,
		// palomamoduletypes.ModuleName,
		// wasm.ModuleName,
		// evmmoduletypes.ModuleName,
		// gravitymoduletypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		// treasurymoduletypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.ModuleManager.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		crisistypes.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		// schedulermoduletypes.ModuleName,
		consensusmoduletypes.ModuleName,
		// valsetmoduletypes.ModuleName,
		// palomamoduletypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		// wasm.ModuleName,
		// evmmoduletypes.ModuleName,
		// gravitymoduletypes.ModuleName,
		// treasurymoduletypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.ModuleManager.RegisterServices(app.configurator)

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.ModuleManager` and `app.configurator` are
	// set.
	// app.RegisterUpgradeHandlers(semverVersion)

	// Create the simulation manager and define the order of the modules for
	// deterministic simulations.
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(
			app.appCodec,
			app.AccountKeeper,
			authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	/*baseAnteHandler*/
	_, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)
	if err != nil {
		panic(err)
	}
	anteDecorators := []sdk.AnteDecorator{
		// palomamodule.NewAnteHandlerDecorator(baseAnteHandler),
	}
	anteDecorators = append(anteDecorators)
	// palomamodule.NewLogMsgDecorator(app.appCodec),
	// 	palomamodule.NewVerifyAuthorisedSignatureDecorator(app.FeeGrantKeeper),

	anteHandler := sdk.ChainAnteDecorators(
		anteDecorators...,
	)

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// app.ScopedWasmKeeper = scopedWasmKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper

	// if mngr := app.SnapshotManager(); mngr != nil {
	// 	err := mngr.RegisterExtensions(wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.wasmKeeper))
	// 	if err != nil {
	// 		panic(fmt.Errorf("failed to register snapshot extension: %w", err))
	// 	}
	// }

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	consensusRegistry.Add(
	// app.EvmKeeper,
	)

	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	// paramsKeeper.Subspace(schedulermoduletypes.ModuleName)
	paramsKeeper.Subspace(consensusmoduletypes.ModuleName)
	// paramsKeeper.Subspace(valsetmoduletypes.ModuleName)
	// paramsKeeper.Subspace(wasm.ModuleName)
	// paramsKeeper.Subspace(evmmoduletypes.ModuleName)
	// paramsKeeper.Subspace(gravitymoduletypes.ModuleName)

	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager.DefaultGenesis(app.appCodec)
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
