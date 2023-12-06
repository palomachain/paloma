package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		Cdc        types.CodecMarshaler
		storeKey   store.KVStoreService
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		ider keeperutil.IDGenerator

		valset types.ValsetKeeper

		registry *registry
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey store.KVStoreService,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	valsetKeeper types.ValsetKeeper,
	reg *registry,
) *Keeper {

	k := &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		valset:     valsetKeeper,
		registry:   reg,
	}
	ider := keeperutil.NewIDGenerator(k, nil)
	k.ider = ider

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Store(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.memKey)
}
