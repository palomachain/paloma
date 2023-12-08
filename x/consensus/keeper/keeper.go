package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
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

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.Logger(sdkCtx).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Store(ctx context.Context) storetypes.KVStore {
	s := runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))

	return prefix.NewStore(s, []byte("store"))
}
