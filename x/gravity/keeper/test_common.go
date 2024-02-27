package keeper

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ccodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	ccrypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	chainparams "github.com/palomachain/paloma/app/params"
	"github.com/palomachain/paloma/testutil/common"
	utilkeeper "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus"
	consensuskeeper "github.com/palomachain/paloma/x/consensus/keeper"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm"
	evmkeeper "github.com/palomachain/paloma/x/evm/keeper"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/palomachain/paloma/x/scheduler"
	"github.com/palomachain/paloma/x/valset"
	valsetkeeper "github.com/palomachain/paloma/x/valset/keeper"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/require"
)

var (
	testERC20Address string = "0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e"
	testDenom        string = "ugrain"

	// ModuleBasics is a mock module basic manager for testing
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distribution.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
				// distrclient.ProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		vesting.AppModuleBasic{},
	)
)

var (
	// ConsPrivKeys generate ed25519 ConsPrivKeys to be used for validator operator keys
	ConsPrivKeys = []ccrypto.PrivKey{
		ed25519.GenPrivKey(),
		ed25519.GenPrivKey(),
		ed25519.GenPrivKey(),
		ed25519.GenPrivKey(),
		ed25519.GenPrivKey(),
	}

	// ConsPubKeys holds the consensus public keys to be used for validator operator keys
	ConsPubKeys = []ccrypto.PubKey{
		ConsPrivKeys[0].PubKey(),
		ConsPrivKeys[1].PubKey(),
		ConsPrivKeys[2].PubKey(),
		ConsPrivKeys[3].PubKey(),
		ConsPrivKeys[4].PubKey(),
	}

	// AccPrivKeys generate secp256k1 pubkeys to be used for account pub keys
	AccPrivKeys = []ccrypto.PrivKey{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	// AccPubKeys holds the pub keys for the account keys
	AccPubKeys = []ccrypto.PubKey{
		AccPrivKeys[0].PubKey(),
		AccPrivKeys[1].PubKey(),
		AccPrivKeys[2].PubKey(),
		AccPrivKeys[3].PubKey(),
		AccPrivKeys[4].PubKey(),
	}

	// AccAddrs holds the sdk.AccAddresses
	AccAddrs = []sdk.AccAddress{
		sdk.AccAddress(AccPubKeys[0].Address()),
		sdk.AccAddress(AccPubKeys[1].Address()),
		sdk.AccAddress(AccPubKeys[2].Address()),
		sdk.AccAddress(AccPubKeys[3].Address()),
		sdk.AccAddress(AccPubKeys[4].Address()),
	}

	// ValAddrs holds the sdk.ValAddresses
	ValAddrs = []sdk.ValAddress{
		sdk.ValAddress(AccPubKeys[0].Address()),
		sdk.ValAddress(AccPubKeys[1].Address()),
		sdk.ValAddress(AccPubKeys[2].Address()),
		sdk.ValAddress(AccPubKeys[3].Address()),
		sdk.ValAddress(AccPubKeys[4].Address()),
	}

	// AccPubKeys holds the pub keys for the account keys
	OrchPubKeys = []ccrypto.PubKey{
		OrchPrivKeys[0].PubKey(),
		OrchPrivKeys[1].PubKey(),
		OrchPrivKeys[2].PubKey(),
		OrchPrivKeys[3].PubKey(),
		OrchPrivKeys[4].PubKey(),
	}

	// Orchestrator private keys
	OrchPrivKeys = []ccrypto.PrivKey{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	// AccAddrs holds the sdk.AccAddresses
	OrchAddrs = []sdk.AccAddress{
		sdk.AccAddress(OrchPubKeys[0].Address()),
		sdk.AccAddress(OrchPubKeys[1].Address()),
		sdk.AccAddress(OrchPubKeys[2].Address()),
		sdk.AccAddress(OrchPubKeys[3].Address()),
		sdk.AccAddress(OrchPubKeys[4].Address()),
	}

	ethPrivKey1, _ = crypto.GenerateKey()
	ethPrivKey2, _ = crypto.GenerateKey()
	ethPrivKey3, _ = crypto.GenerateKey()
	ethPrivKey4, _ = crypto.GenerateKey()
	ethPrivKey5, _ = crypto.GenerateKey()
	EthPrivKeys    = []*ecdsa.PrivateKey{
		ethPrivKey1,
		ethPrivKey2,
		ethPrivKey3,
		ethPrivKey4,
		ethPrivKey5,
	}

	// EthAddrs holds etheruem addresses
	EthAddrs = []gethcommon.Address{
		gethcommon.BytesToAddress(crypto.PubkeyToAddress(EthPrivKeys[0].PublicKey).Bytes()),
		gethcommon.BytesToAddress(crypto.PubkeyToAddress(EthPrivKeys[1].PublicKey).Bytes()),
		gethcommon.BytesToAddress(crypto.PubkeyToAddress(EthPrivKeys[2].PublicKey).Bytes()),
		gethcommon.BytesToAddress(crypto.PubkeyToAddress(EthPrivKeys[3].PublicKey).Bytes()),
		gethcommon.BytesToAddress(crypto.PubkeyToAddress(EthPrivKeys[4].PublicKey).Bytes()),
	}

	// TokenContractAddrs holds example token contract addresses
	TokenContractAddrs = []string{
		"0x6b175474e89094c44da98b954eedeac495271d0f", // DAI
		"0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", // YFI
		"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984", // UNI
		"0xc00e94cb662c3520282e6f5717214004a7f26888", // COMP
		"0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f", // SNX
	}

	// InitTokens holds the number of tokens to initialize an account with
	InitTokens = sdk.TokensFromConsensusPower(110, sdk.DefaultPowerReduction)

	// InitCoins holds the number of coins to initialize an account with
	InitCoins = sdk.NewCoins(sdk.NewCoin(TestingStakeParams.BondDenom, InitTokens))

	// StakingAmount holds the staking power to start a validator with
	StakingAmount = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)

	// StakingCoins holds the staking coins to start a validator with
	StakingCoins = sdk.NewCoins(sdk.NewCoin(TestingStakeParams.BondDenom, StakingAmount))

	// TestingStakeParams is a set of staking params for testing
	TestingStakeParams = stakingtypes.Params{
		UnbondingTime:     100,
		MaxValidators:     10,
		MaxEntries:        10,
		HistoricalEntries: 10000,
		BondDenom:         testDenom,
		MinCommissionRate: math.LegacyNewDecWithPrec(5, 2),
	}

	// TestingGravityParams is a set of gravity params for testing
	TestingGravityParams = types.Params{
		ContractSourceHash:           "62328f7bc12efb28f86111d08c29b39285680a906ea0e524e0209d6f6657b713",
		BridgeEthereumAddress:        "0x8858eeb3dfffa017d4bce9801d340d36cf895ccf",
		BridgeChainId:                11,
		SignedBatchesWindow:          10,
		TargetBatchTimeout:           60001,
		AverageBlockTime:             5000,
		AverageEthereumBlockTime:     15000,
		SlashFractionBatch:           math.LegacyNewDecWithPrec(1, 2),
		SlashFractionBadEthSignature: math.LegacyNewDecWithPrec(1, 2),
	}
)

// TestInput stores the various keepers required to test gravity
type TestInput struct {
	GravityKeeper     Keeper
	AccountKeeper     authkeeper.AccountKeeper
	StakingKeeper     stakingkeeper.Keeper
	ValsetKeeper      valsetkeeper.Keeper
	SlashingKeeper    slashingkeeper.Keeper
	DistKeeper        distrkeeper.Keeper
	BankKeeper        bankkeeper.BaseKeeper
	GovKeeper         govkeeper.Keeper
	IbcTransferKeeper ibctransferkeeper.Keeper
	Context           context.Context
	Marshaler         codec.Codec
	LegacyAmino       *codec.LegacyAmino
	GravityStoreKey   *storetypes.KVStoreKey
}

func addValidators(t *testing.T, input *TestInput, count int) {
	// Initialize each of the validators
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(&input.StakingKeeper)
	for i := 0; i < count; i++ {

		// Initialize the account for the key
		acc := input.AccountKeeper.NewAccount(
			input.Context,
			authtypes.NewBaseAccount(AccAddrs[i], AccPubKeys[i], uint64(i), 0),
		)

		// Set the balance for the account
		require.NoError(t, input.BankKeeper.MintCoins(input.Context, types.ModuleName, InitCoins))
		require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(input.Context, types.ModuleName, acc.GetAddress(), InitCoins))

		// Set the account in state
		input.AccountKeeper.SetAccount(input.Context, acc)

		// Create a validator for that account using some of the tokens in the account
		// and the staking handler
		_, err := stakingMsgSvr.CreateValidator(sdk.UnwrapSDKContext(input.Context), NewTestMsgCreateValidator(ValAddrs[i], ConsPubKeys[i], StakingAmount))

		// Return error if one exists
		require.NoError(t, err)
	}

	// Run the staking endblocker to ensure valset is correct in state
	_, err := input.StakingKeeper.EndBlocker(input.Context)
	require.NoError(t, err)

	// Register eth addresses and orchestrator address for each validator
	for i, addr := range ValAddrs {
		ethAddr, err := types.NewEthAddress(EthAddrs[i].String())
		require.NoError(t, err)

		validator, found := input.StakingKeeper.GetValidator(input.Context, addr)
		require.NoError(t, found)

		pubKey, err := validator.ConsPubKey()
		require.NoError(t, err)
		valAddress, err := utilkeeper.ValAddressFromBech32(input.GravityKeeper.AddressCodec, validator.GetOperator())
		if err != nil {
			require.NoError(t, err)
		}
		sdkCtx := sdk.UnwrapSDKContext(input.Context)
		err = input.ValsetKeeper.AddExternalChainInfo(sdkCtx, valAddress, []*valsettypes.ExternalChainInfo{
			{
				ChainType:        "evm",
				ChainReferenceID: "test-chain",
				Address:          ethAddr.GetAddress().String(),
				Pubkey:           pubKey.Bytes(),
			},
		})
		require.NoError(t, err)
	}

	// Create a Snapshot
	sdkCtx := sdk.UnwrapSDKContext(input.Context)
	_, err = input.ValsetKeeper.TriggerSnapshotBuild(sdkCtx)
	require.NoError(t, err)
}

// SetupFiveValChain does all the initialization for a 5 Validator chain using the keys here
func SetupFiveValChain(t *testing.T) (TestInput, context.Context) {
	t.Helper()
	input := CreateTestEnv(t)

	// Set the params for our modules
	err := input.StakingKeeper.SetParams(input.Context, TestingStakeParams)
	require.NoError(t, err)

	addValidators(t, &input, 5)

	// Return the test input
	return input, input.Context
}

// SetupTestChain sets up a test environment with the provided validator voting weights
func SetupTestChain(t *testing.T, weights []uint64) (TestInput, context.Context) {
	t.Helper()
	input, ctx := SetupFiveValChain(t)

	// Set the params for our modules
	TestingStakeParams.MaxValidators = 100
	err := input.StakingKeeper.SetParams(ctx, TestingStakeParams)
	require.NoError(t, err)

	// Initialize each of the validators
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(&input.StakingKeeper)
	for i, weight := range weights {
		consPrivKey := ed25519.GenPrivKey()
		consPubKey := consPrivKey.PubKey()
		valPrivKey := secp256k1.GenPrivKey()
		valPubKey := valPrivKey.PubKey()
		valAddr := sdk.ValAddress(valPubKey.Address())
		accAddr := sdk.AccAddress(valPubKey.Address())

		// Initialize the account for the key
		acc := input.AccountKeeper.NewAccount(
			ctx,
			authtypes.NewBaseAccount(accAddr, valPubKey, uint64(i), 0),
		)

		// Set the balance for the account
		weightCoins := sdk.NewCoins(sdk.NewInt64Coin(TestingStakeParams.BondDenom, int64(weight)))
		require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, weightCoins))
		require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddr, weightCoins))

		// Set the account in state
		input.AccountKeeper.SetAccount(ctx, acc)

		// Create a validator for that account using some of the tokens in the account
		// and the staking handler
		_, err := stakingMsgSvr.CreateValidator(sdk.UnwrapSDKContext(input.Context), NewTestMsgCreateValidator(valAddr, consPubKey, math.NewIntFromUint64(weight)))

		require.NoError(t, err)

		// Run the staking endblocker to ensure valset is correct in state
		_, err = input.StakingKeeper.EndBlocker(input.Context)
		require.NoError(t, err)

		// increase block height by 100 blocks
		sdkCtx := sdk.UnwrapSDKContext(input.Context)
		input.Context = sdkCtx.WithBlockHeight(sdkCtx.BlockHeight() + 100)

		// Run the staking endblocker to ensure valset is correct in state
		_, err = input.StakingKeeper.EndBlocker(input.Context)
		require.NoError(t, err)

	}

	// some inputs can cause the validator creation ot not work, this checks that
	// everything was successful.  Adding 5 for the 5 validator initial setup
	validators, err := input.StakingKeeper.GetBondedValidatorsByPower(input.Context)
	require.NoError(t, err)
	require.Equal(t, len(weights)+5, len(validators))

	// Return the test input
	return input, input.Context
}

// CreateTestEnv creates the keeper testing environment for gravity
func CreateTestEnv(t *testing.T) TestInput {
	t.Helper()

	common.SetupPalomaPrefixes()
	// Initialize store keys
	gravityKey := storetypes.NewKVStoreKey(types.StoreKey)
	keyAcc := storetypes.NewKVStoreKey(authtypes.StoreKey)
	keyStaking := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	keyBank := storetypes.NewKVStoreKey(banktypes.StoreKey)
	keyDistro := storetypes.NewKVStoreKey(distrtypes.StoreKey)
	keyParams := storetypes.NewKVStoreKey(paramstypes.StoreKey)
	tkeyParams := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	keyGov := storetypes.NewKVStoreKey(govtypes.StoreKey)
	keySlashing := storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	keyCapability := storetypes.NewKVStoreKey(capabilitytypes.StoreKey)
	keyUpgrade := storetypes.NewKVStoreKey(upgradetypes.StoreKey)
	keyIbc := storetypes.NewKVStoreKey(ibcexported.StoreKey)
	keyIbcTransfer := storetypes.NewKVStoreKey(ibctransfertypes.StoreKey)

	keyValset := storetypes.NewKVStoreKey(valsettypes.StoreKey)
	memKeyValset := storetypes.NewKVStoreKey(valsettypes.MemStoreKey)
	keyConsensus := storetypes.NewKVStoreKey(consensustypes.StoreKey)
	memKeyConsensus := storetypes.NewKVStoreKey(consensustypes.MemStoreKey)
	keyEvm := storetypes.NewKVStoreKey(evmtypes.StoreKey)
	memKeyEvm := storetypes.NewKVStoreKey(evmtypes.MemStoreKey)
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
	// Initialize memory database and mount stores on it
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	ms.MountStoreWithDB(gravityKey, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyStaking, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyDistro, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, storetypes.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyGov, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySlashing, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyCapability, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyUpgrade, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIbc, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIbcTransfer, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyValset, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(memKeyValset, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyConsensus, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(memKeyConsensus, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyEvm, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(memKeyEvm, storetypes.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	// Create context.Context
	ctx := sdk.NewContext(ms, tmproto.Header{
		Version: tmversion.Consensus{
			Block: 0,
			App:   0,
		},
		ChainID: "",
		Height:  1234567,
		Time:    time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
		LastBlockId: tmproto.BlockID{
			Hash: []byte{},
			PartSetHeader: tmproto.PartSetHeader{
				Total: 0,
				Hash:  []byte{},
			},
		},
		LastCommitHash:     []byte{},
		DataHash:           []byte{},
		ValidatorsHash:     []byte{},
		NextValidatorsHash: []byte{},
		ConsensusHash:      []byte{},
		AppHash:            []byte{},
		LastResultsHash:    []byte{},
		EvidenceHash:       []byte{},
		ProposerAddress:    []byte{},
	}, false, log.NewTestLogger(t))

	appCodec := codec.NewProtoCodec(cdc.InterfaceRegistry())
	legacyAmino := codec.NewLegacyAmino()

	marshaler := MakeTestMarshaler()

	paramsKeeper := paramskeeper.NewKeeper(marshaler, legacyAmino, keyParams, tkeyParams)
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(types.DefaultParamspace)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(valsettypes.ModuleName)
	paramsKeeper.Subspace(consensustypes.ModuleName)
	paramsKeeper.Subspace(evmtypes.ModuleName)

	// this is also used to initialize module accounts for all the map keys
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		types.ModuleName:               {authtypes.Minter, authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keyAcc),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(chainparams.AccountAddressPrefix),
		chainparams.AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	blockedAddr := make(map[string]bool, len(maccPerms))
	for acc := range maccPerms {
		blockedAddr[authtypes.NewModuleAddress(acc).String()] = true
	}
	appCodec = codec.NewProtoCodec(cdc.InterfaceRegistry())
	bankKeeper := bankkeeper.NewBaseKeeper(
		marshaler,
		runtime.NewKVStoreService(keyBank),
		accountKeeper,
		blockedAddr,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)

	err = bankKeeper.SetParams(ctx, banktypes.Params{
		SendEnabled:        []*banktypes.SendEnabled{},
		DefaultSendEnabled: true,
	})

	require.NoError(t, err)

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keyStaking),
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(chainparams.ValidatorAddressPrefix),
		authcodec.NewBech32Codec(chainparams.ConsNodeAddressPrefix),
	)

	err = stakingKeeper.SetParams(ctx, TestingStakeParams)
	require.NoError(t, err)

	distKeeper := distrkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keyDistro),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	err = distKeeper.Params.Set(ctx, distrtypes.DefaultParams())
	require.NoError(t, err)

	// set genesis items required for distribution
	err = distKeeper.FeePool.Set(ctx, distrtypes.InitialFeePool())
	require.NoError(t, err)

	// set up initial accounts
	for name := range maccPerms {
		mod, _ := accountKeeper.GetModuleAccountAndPermissions(ctx, name)
		// mod := authtypes.NewEmptyModuleAccount(name, perms...)
		if name == distrtypes.ModuleName {
			// some big pot to pay out
			amt := sdk.NewCoins(sdk.NewInt64Coin(testDenom, 500000))
			err = bankKeeper.MintCoins(ctx, types.ModuleName, amt)
			require.NoError(t, err)
			err = bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, mod.GetName(), amt)
			require.NoError(t, err)
			// distribution module balance must be outstanding rewards + community pool in order to pass
			// invariants checks, therefore we must add any amount we add to the module balance to the fee pool
			feePool, err := distKeeper.FeePool.Get(ctx)
			require.NoError(t, err)
			newCoins := feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amt...)...)
			feePool.CommunityPool = newCoins

			// accountKeeper.SetModuleAccount(ctx, mod)
			err = distKeeper.FeePool.Set(ctx, feePool)
			require.NoError(t, err)
		}
	}
	stakeAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	moduleAcct := accountKeeper.GetAccount(ctx, stakeAddr)
	require.NotNil(t, moduleAcct)
	bApp := *baseapp.NewBaseApp("test", log.NewTestLogger(t), db, MakeTestEncodingConfig().TxConfig.TxDecoder())
	govKeeper := govkeeper.NewKeeper(
		marshaler,
		runtime.NewKVStoreService(keyGov),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distKeeper,
		bApp.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	slashingKeeper := slashingkeeper.NewKeeper(
		marshaler,
		codec.NewLegacyAmino(),
		runtime.NewKVStoreService(keySlashing),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	upgradeKeeper := upgradekeeper.NewKeeper(
		make(map[int64]bool),
		runtime.NewKVStoreService(keyUpgrade),
		marshaler,
		"",
		&bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	capabilityKeeper := *capabilitykeeper.NewKeeper(
		marshaler,
		keyCapability,
		memKeys[capabilitytypes.MemStoreKey],
	)

	scopedIbcKeeper := capabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	ibcKeeper := *ibckeeper.NewKeeper(
		marshaler,
		keyIbc,
		getSubspace(paramsKeeper, ibcexported.ModuleName),
		stakingKeeper,
		upgradeKeeper,
		scopedIbcKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	scopedTransferKeeper := capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	ibcTransferKeeper := ibctransferkeeper.NewKeeper(
		marshaler,
		keyIbcTransfer,
		getSubspace(paramsKeeper, ibctransfertypes.ModuleName),
		ibcKeeper.ChannelKeeper,
		ibcKeeper.ChannelKeeper,
		ibcKeeper.PortKeeper,
		accountKeeper,
		bankKeeper,
		scopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	valsetKeeper := valsetkeeper.NewKeeper(
		marshaler,
		runtime.NewKVStoreService(keyValset),
		getSubspace(paramsKeeper, valsettypes.ModuleName),
		stakingKeeper,
		slashingKeeper,
		"v1.5.0",
		sdk.DefaultPowerReduction,
		authcodec.NewBech32Codec(chainparams.ValidatorAddressPrefix),
	)

	consensusRegistry := consensuskeeper.NewRegistry()
	consensusKeeper := consensuskeeper.NewKeeper(
		marshaler,
		runtime.NewKVStoreService(keyConsensus),
		getSubspace(paramsKeeper, consensustypes.ModuleName),
		valsetKeeper,
		consensusRegistry,
	)

	evmKeeper := evmkeeper.NewKeeper(
		marshaler,
		runtime.NewKVStoreService(keyEvm),
		consensusKeeper,
		valsetKeeper,
		authcodec.NewBech32Codec(chainparams.ValidatorAddressPrefix),
	)

	valsetKeeper.EvmKeeper = evmKeeper

	err = evmKeeper.AddSupportForNewChain(
		ctx,
		"test-chain",
		1,
		uint64(123),
		"0x1234",
		big.NewInt(55),
	)
	require.NoError(t, err)

	k := NewKeeper(
		marshaler,
		accountKeeper,
		stakingKeeper,
		bankKeeper,
		slashingKeeper,
		distKeeper,
		ibcTransferKeeper,
		evmKeeper,
		NewGravityStoreGetter(gravityKey),
		"",
		authcodec.NewBech32Codec(chainparams.ValidatorAddressPrefix),
	)

	stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			distKeeper.Hooks(),
			slashingKeeper.Hooks(),
			k.Hooks(),
		),
	)

	// set gravityIDs for batches and tx items, simulating genesis setup
	err = k.setLastObservedEventNonce(ctx, 0)
	require.NoError(t, err)
	err = k.SetLastSlashedBatchBlock(ctx, 0)
	require.NoError(t, err)
	k.setID(ctx, 0, types.KeyLastTXPoolID)
	k.setID(ctx, 0, types.KeyLastOutgoingBatchID)

	k.SetParams(ctx, TestingGravityParams)

	// Add ERC20 to Denom mapping
	ethAddr, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)
	err = k.setDenomToERC20(ctx, "test-chain", testDenom, *ethAddr)
	require.NoError(t, err)

	//// Add some validators
	//validators := testutil.GenValidators(5, 5000)
	//for _, val := range validators {
	//	stakingKeeper.SetValidator(ctx, val)
	//}
	//
	//for _, validator := range validators {
	//	valAddr, err := validator.GetConsAddr()
	//	require.NoError(t, err)
	//	pubKey, err := validator.ConsPubKey()
	//	require.NoError(t, err)
	//	err = valsetKeeper.AddExternalChainInfo(ctx, validator.GetOperator(), []*valsettypes.ExternalChainInfo{
	//		{
	//			ChainType:        "evm",
	//			ChainReferenceID: "test-chain",
	//			Address:          valAddr.String(),
	//			Pubkey:           pubKey.Bytes(),
	//		},
	//	})
	//	require.NoError(t, err)
	//
	//	distKeeper.SetValidatorHistoricalRewards(
	//		ctx,
	//		validator.GetOperator(),
	//		0,
	//		distrtypes.ValidatorHistoricalRewards{
	//			ReferenceCount: 1,
	//		},
	//	)
	//}
	//
	//// Create a Snapshot
	//_, err = valsetKeeper.TriggerSnapshotBuild(ctx)
	//require.NoError(t, err)

	testInput := TestInput{
		GravityKeeper:     k,
		AccountKeeper:     accountKeeper,
		StakingKeeper:     *stakingKeeper,
		SlashingKeeper:    slashingKeeper,
		ValsetKeeper:      *valsetKeeper,
		DistKeeper:        distKeeper,
		BankKeeper:        bankKeeper,
		GovKeeper:         *govKeeper,
		IbcTransferKeeper: ibcTransferKeeper,
		Context:           ctx,
		Marshaler:         marshaler,
		LegacyAmino:       legacyAmino,
	}

	sdkCtx := sdk.UnwrapSDKContext(testInput.Context)
	// check invariants before starting
	sdkCtx.Logger().Info("Asserting invariants on new test env")
	testInput.AssertInvariants()
	return testInput
}

// AssertInvariants tests each modules invariants individually, this is easier than
// dealing with all the init required to get the crisis keeper working properly by
// running appModuleBasic for every module and allowing them to register their invariants
func (t TestInput) AssertInvariants() {
	distrInvariantFunc := distrkeeper.AllInvariants(t.DistKeeper)
	bankInvariantFunc := bankkeeper.AllInvariants(t.BankKeeper)
	govInvariantFunc := govkeeper.ModuleAccountInvariant(&t.GovKeeper, t.BankKeeper)
	stakeInvariantFunc := stakingkeeper.AllInvariants(&t.StakingKeeper)
	gravInvariantFunc := AllInvariants(t.GravityKeeper)
	sdkCtx := sdk.UnwrapSDKContext(t.Context)
	invariantStr, invariantViolated := distrInvariantFunc(sdkCtx)
	if invariantViolated {
		panic(invariantStr)
	}
	invariantStr, invariantViolated = bankInvariantFunc(sdkCtx)
	if invariantViolated {
		panic(invariantStr)
	}
	invariantStr, invariantViolated = govInvariantFunc(sdkCtx)
	if invariantViolated {
		panic(invariantStr)
	}
	invariantStr, invariantViolated = stakeInvariantFunc(sdkCtx)
	if invariantViolated {
		panic(invariantStr)
	}
	invariantStr, invariantViolated = gravInvariantFunc(sdkCtx)
	if invariantViolated {
		panic(invariantStr)
	}

	sdkCtx.Logger().Info("All invariants successful")
}

// getSubspace returns a param subspace for a given module name.
func getSubspace(k paramskeeper.Keeper, moduleName string) paramstypes.Subspace {
	subspace, _ := k.GetSubspace(moduleName)
	return subspace
}

// MakeTestCodec creates a legacy amino codec for testing
func MakeTestCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	auth.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	bank.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	staking.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	distribution.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	sdk.RegisterLegacyAminoCodec(cdc)
	ccodec.RegisterCrypto(cdc)
	params.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	types.RegisterCodec(cdc)
	return cdc
}

// MakeTestMarshaler creates a proto codec for use in testing
func MakeTestMarshaler() codec.Codec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	types.RegisterInterfaces(interfaceRegistry)
	return codec.NewProtoCodec(interfaceRegistry)
}

func MakeTestEncodingConfig() chainparams.EncodingConfig {
	encodingConfig := chainparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

// MintVouchersFromAir creates new gravity vouchers given erc20tokens
func MintVouchersFromAir(t *testing.T, ctx context.Context, k Keeper, dest sdk.AccAddress, amount types.InternalERC20Token) sdk.Coin {
	coin := sdk.NewCoin(testDenom, amount.Amount)
	vouchers := sdk.Coins{coin}
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, vouchers)
	require.NoError(t, err)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, dest, vouchers)
	require.NoError(t, err)
	return coin
}

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey ccrypto.PubKey, amt math.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(2, 1), math.LegacyNewDecWithPrec(5, 3))
	out, err := stakingtypes.NewMsgCreateValidator(
		address.String(), pubKey, sdk.NewCoin(testDenom, amt),
		stakingtypes.Description{
			Moniker:         "xyz",
			Identity:        "1",
			Website:         "",
			SecurityContact: "",
			Details:         "",
		}, commission, math.OneInt(),
	)
	if err != nil {
		panic(err)
	}
	return out
}

func NewTestMsgUnDelegateValidator(address sdk.ValAddress, amt math.Int) *stakingtypes.MsgUndelegate {
	msg := stakingtypes.NewMsgUndelegate(sdk.AccAddress(address).String(), address.String(), sdk.NewCoin(testDenom, amt))
	return msg
}
