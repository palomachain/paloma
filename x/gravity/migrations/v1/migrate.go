package v1

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/exported"
	"github.com/palomachain/paloma/x/gravity/types"
)

// migrateParams will set the params to store from subspace

func MigrateParams(
	ctx sdk.Context,
	store storetypes.KVStore,
	subspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	var currParams types.Params
	subspace.GetParamSet(ctx, &currParams)
	if err := currParams.ValidateBasic(); err != nil {
		return err
	}
	bz, err := cdc.Marshal(&currParams)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}
