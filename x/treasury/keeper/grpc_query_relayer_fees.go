package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/treasury/types"
)

// QueryRelayerFee implements types.QueryServer.
func (Keeper) QueryRelayerFee(context.Context, *types.QueryRelayerFeeRequest) (*types.RelayerFee, error) {
	panic("unimplemented")
}

// QueryRelayerFees implements types.QueryServer.
func (Keeper) QueryRelayerFees(context.Context, *types.Empty) (*types.QueryRelayerFeesResponse, error) {
	panic("unimplemented")
}
