package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/treasury/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := common.SdkContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}
