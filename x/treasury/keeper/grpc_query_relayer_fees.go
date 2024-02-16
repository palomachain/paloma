package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/types"
)

// QueryRelayerFee implements types.QueryServer.
func (k Keeper) QueryRelayerFee(ctx context.Context, req *types.QueryRelayerFeeRequest) (*types.RelayerFee, error) {
	a, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	return k.getRelayerFee(ctx, a)
}

// QueryRelayerFees implements types.QueryServer.
func (k Keeper) QueryRelayerFees(ctx context.Context, _ *types.Empty) (*types.QueryRelayerFeesResponse, error) {
	fees, err := k.getRelayerFees(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryRelayerFeesResponse{
		Fees: fees,
	}, nil
}
