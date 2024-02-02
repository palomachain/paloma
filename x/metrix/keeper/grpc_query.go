package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/metrix/types"
)

var _ types.QueryServer = Keeper{}

// Validator implements types.QueryServer.
func (k Keeper) Validator(goCtx context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValAddress)
	if err != nil {
		return nil, fmt.Errorf("query validator: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	metrics, err := k.GetValidatorMetrics(ctx, valAddr)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	if metrics == nil {
		return nil, fmt.Errorf("no metrics found")
	}

	return &types.QueryValidatorResponse{
		ValMetrics: *metrics,
	}, nil
}

// Validators implements types.QueryServer.
func (k Keeper) Validators(goCtx context.Context, _ *types.Empty) (*types.QueryValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	metrics := make([]types.ValidatorMetrics, 0, 200)
	if err := k.metrics.Iterate(ctx, func(_ []byte, m *types.ValidatorMetrics) bool {
		if m != nil {
			metrics = append(metrics, *m)
		}

		return true
	}); err != nil {
		return nil, fmt.Errorf("iterate metrics: %w", err)
	}

	return &types.QueryValidatorsResponse{
		ValMetrics: metrics,
	}, nil
}

// HistoricRelayData implements types.QueryServer.
func (k Keeper) HistoricRelayData(goCtx context.Context, req *types.QueryHistoricRelayDataRequest) (*types.QueryHistoricRelayDataResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValAddress)
	if err != nil {
		return nil, fmt.Errorf("query validator: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	history, err := k.GetValidatorHistory(ctx, valAddr)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	if history == nil {
		return nil, fmt.Errorf("no history found")
	}

	return &types.QueryHistoricRelayDataResponse{
		History: *history,
	}, nil
}
