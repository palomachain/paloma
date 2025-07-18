package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/valset/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	defer func() {
		if r := recover(); r != nil {
			params = types.DefaultParams()
		}
	}()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramstore.GetParamSet(sdkCtx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramstore.SetParamSet(sdkCtx, &params)
}
