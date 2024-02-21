package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/scheduler/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryGetJobByID(goCtx context.Context, req *types.QueryGetJobByIDRequest) (*types.QueryGetJobByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := common.SdkContext(goCtx)

	job, err := k.GetJob(ctx, req.GetJobID())
	if err != nil {
		return nil, err
	}

	return &types.QueryGetJobByIDResponse{
		Job: job,
	}, nil
}
