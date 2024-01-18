package v1

import (
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/exported"
	"github.com/palomachain/paloma/x/gravity/types"
)

func MigrateParams(
	ctx sdk.Context,
	storeService store.KVStoreService,
	subspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	store := storeService.OpenKVStore(ctx)
	var currParams types.Params
	subspace.GetParamSet(ctx, &currParams)
	if err := currParams.ValidateBasic(); err != nil {
		return err
	}
	bz, err := cdc.Marshal(&currParams)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}
