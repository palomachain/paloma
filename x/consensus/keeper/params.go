package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) types.Params {
	return types.NewParams()
}

// SetParams set the params
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramstore.SetParamSet(sdkCtx, &params)
}
