package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RelayerFee(ctx context.Context, req *types.QueryRelayerFeeRequest) (*types.RelayerFeeSetting, error) {
	addr, err := sdk.ValAddressFromBech32(req.ValAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse validator address: %w", err)
	}

	f, err := k.relayerFees.Get(sdk.UnwrapSDKContext(ctx), addr)
	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			return &types.RelayerFeeSetting{}, nil
		}
		return nil, err
	}

	return f, nil
}

func (k Keeper) RelayerFees(ctx context.Context, req *types.QueryRelayerFeesRequest) (*types.QueryRelayerFeesResponse, error) {
	i, err := k.GetRelayerFeesByChainReferenceID(ctx, req.ChainReferenceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get relayer fees: %w", err)
	}

	response := make([]types.RelayerFeeSetting, 0, len(i))
	for k, v := range i {
		response = append(response, types.RelayerFeeSetting{
			ValAddress: k,
			Fees: []types.RelayerFeeSetting_FeeSetting{
				{
					Multiplicator:    v,
					ChainReferenceId: req.ChainReferenceId,
				},
			},
		})
	}

	return &types.QueryRelayerFeesResponse{RelayerFees: response}, nil
}
