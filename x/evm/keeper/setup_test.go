package keeper

import (
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/evm/types/mocks"
	"github.com/stretchr/testify/require"
)

type mockedServices struct {
	ConsensusKeeper *mocks.ConsensusKeeper
	ValsetKeeper    *mocks.ValsetKeeper
	MsgSender       *mocks.MsgSender
}

func NewEvmKeeper(t testutil.TB) (*Keeper, mockedServices, sdk.Context) {
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
		"EvmParams",
	)

	ms := mockedServices{
		ConsensusKeeper: mocks.NewConsensusKeeper(t),
		ValsetKeeper:    mocks.NewValsetKeeper(t),
		MsgSender:       mocks.NewMsgSender(t),
	}
	k := NewKeeper(
		appCodec,
		storeKey,
		memStoreKey,
		paramsSubspace,
		ms.ConsensusKeeper,
		ms.ValsetKeeper,
	)

	k.msgSender = ms.MsgSender

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ms, ctx
}
