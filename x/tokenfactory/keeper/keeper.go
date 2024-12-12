package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

type (
	Keeper struct {
		storeKey storetypes.StoreKey

		paramSpace paramtypes.Subspace

		accountKeeper       types.AccountKeeper
		bankKeeper          types.BankKeeper
		communityPoolKeeper types.CommunityPoolKeeper

		authority string
	}
)

func NewKeeper(
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	communityPoolKeeper types.CommunityPoolKeeper,
	authority string,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,

		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		communityPoolKeeper: communityPoolKeeper,
		authority:           authority,
	}
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return liblog.FromSDKLogger(sdkCtx.Logger()).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetDenomPrefixStore(ctx context.Context, denom string) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

func (k Keeper) GetCreatorPrefixStore(ctx context.Context, creator string) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

func (k Keeper) GetCreatorsPrefixStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorsPrefix())
}

func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramSpace.GetParamSet(sdkCtx, &params)
	return params
}

func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramSpace.SetParamSet(sdkCtx, &params)
}
