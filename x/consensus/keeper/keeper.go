package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/libcons"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/consensus/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   store.KVStoreService
		paramstore paramtypes.Subspace

		ider keeperutil.IDGenerator

		valset types.ValsetKeeper

		registry         *registry
		evmKeeper        types.EvmKeeper
		consensusChecker *libcons.ConsensusChecker
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey store.KVStoreService,
	ps paramtypes.Subspace,
	valsetKeeper types.ValsetKeeper,
	reg *registry,
) *Keeper {
	k := &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramstore: ps,
		valset:     valsetKeeper,
		registry:   reg,
	}
	ider := keeperutil.NewIDGenerator(k, nil)
	k.ider = ider
	k.consensusChecker = libcons.New(k.valset.GetCurrentSnapshot, k.cdc)

	return k
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return liblog.FromSDKLogger(sdkCtx.Logger()).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Store(ctx context.Context) storetypes.KVStore {
	return runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
}

func (k Keeper) ModuleName() string {
	return types.ModuleName
}

func (k *Keeper) LateInject(evm types.EvmKeeper) {
	k.evmKeeper = evm
}
