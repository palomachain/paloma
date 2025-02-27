package keeper

import (
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
	"github.com/palomachain/paloma/v2/app/params"
	"github.com/palomachain/paloma/v2/testutil"
	"github.com/palomachain/paloma/v2/x/evm/types"
	"github.com/palomachain/paloma/v2/x/evm/types/mocks"
	"github.com/stretchr/testify/require"
)

type mockedServices struct {
	ConsensusKeeper *mocks.ConsensusKeeper
	ValsetKeeper    *mocks.ValsetKeeper
	MsgSender       *mocks.MsgSender
	SkywayKeeper    *mocks.SkywayKeeper
	TreasuryKeeper  *mocks.TreasuryKeeper
	MetrixKeeper    *mocks.MetrixKeeper
}

func NewEvmKeeper(t testutil.TB) (*Keeper, mockedServices, sdk.Context) {
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
		ConsensusKeeper: mocks.NewConsensusKeeper(t),
		ValsetKeeper:    mocks.NewValsetKeeper(t),
		MsgSender:       mocks.NewMsgSender(t),
		SkywayKeeper:    mocks.NewSkywayKeeper(t),
		TreasuryKeeper:  mocks.NewTreasuryKeeper(t),
		MetrixKeeper:    mocks.NewMetrixKeeper(t),
	}
	k := NewKeeper(
		appCodec,
		storeService,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		ms.ConsensusKeeper,
		ms.ValsetKeeper,
		authcodec.NewBech32Codec(params.ValidatorAddressPrefix),
		ms.MetrixKeeper,
		ms.TreasuryKeeper,
	)

	k.msgSender = ms.MsgSender
	k.Skyway = ms.SkywayKeeper

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
