package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/x/skyway/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetBridgeTransferLimits(
	goCtx context.Context,
	_ *emptypb.Empty,
) (*types.QueryBridgeTransferLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	limits, err := k.AllBridgeTransferLimits(ctx)
	if err != nil {
		return nil, err
	}

	res := &types.QueryBridgeTransferLimitsResponse{
		Limits: make([]*types.QueryBridgeTransferLimitsResponse_LimitUsage, len(limits)),
	}

	for i := range limits {
		usage, err := k.BridgeTransferUsage(ctx, limits[i].Token)
		if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
			return nil, err
		}

		res.Limits[i] = &types.QueryBridgeTransferLimitsResponse_LimitUsage{
			Limit: limits[i],
			Usage: usage,
		}
	}

	return res, nil
}
