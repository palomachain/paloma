package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetValidatorAliveUntil(goCtx context.Context, req *types.QueryGetValidatorAliveUntilRequest) (*types.QueryGetValidatorAliveUntilResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := common.SdkContext(goCtx)

	aliveUntil, err := k.ValidatorAliveUntil(ctx, req.GetValAddress())
	if err != nil {
		return nil, err
	}

	return &types.QueryGetValidatorAliveUntilResponse{
		AliveUntilBlockHeight: aliveUntil,
	}, nil
}
