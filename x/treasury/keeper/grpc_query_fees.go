package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/treasury/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryFees(c context.Context, req *types.QueryFeesRequest) (*types.Fees, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := common.SdkContext(c)
	return k.GetFees(ctx)
}
