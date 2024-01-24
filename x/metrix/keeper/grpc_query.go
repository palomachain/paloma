package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/metrix/types"
)

var _ types.QueryServer = Keeper{}

// Validator implements types.QueryServer.
func (Keeper) Validator(context.Context, *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	panic("unimplemented")
}
