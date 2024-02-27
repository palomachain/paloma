package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/scheduler/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	return types.NewParams()
}

// SetParams sets the parameters in the store
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
}
