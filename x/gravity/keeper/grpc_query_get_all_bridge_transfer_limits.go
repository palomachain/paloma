package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetAllBridgeTransferLimits(
	goCtx context.Context,
	_ *emptypb.Empty,
) (*types.QueryAllBridgeTransferLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	limits, err := k.AllBridgeTransferLimits(ctx)
	if err != nil {
		return nil, err
	}

	res := &types.QueryAllBridgeTransferLimitsResponse{
		Limits: make([]*types.BridgeTransferLimitUsage, len(limits)),
	}

	for i := range limits {
		usage, err := k.BridgeTransferUsage(ctx, limits[i].Token)
		if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
			return nil, err
		}

		res.Limits[i] = &types.BridgeTransferLimitUsage{
			Limit: limits[i],
			Usage: usage,
		}
	}

	return res, nil
}
