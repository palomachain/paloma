package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/scheduler/types"
)

func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	return types.NewParams()
}

// SetParams sets the parameters in the store
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
}
