package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/paloma/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) types.Params {
	return types.NewParams()
}

// SetParams set the params
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
}
