package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/valset/types"
)

func (k Keeper) GetPigeonRequirements(goCtx context.Context, req *types.QueryGetPigeonRequirementsRequest) (*types.QueryGetPigeonRequirementsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	pigeonReq, err := k.PigeonRequirements(ctx)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			return nil, err
		}

		pigeonReq = nil
	}

	scheduledReq, err := k.ScheduledPigeonRequirements(ctx)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			return nil, err
		}

		scheduledReq = nil
	}

	return &types.QueryGetPigeonRequirementsResponse{
		PigeonRequirements:          pigeonReq,
		ScheduledPigeonRequirements: scheduledReq,
	}, nil
}
