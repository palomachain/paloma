package keeper

import (
	"context"

	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetAllQueueNames(goCtx context.Context, req *types.QueryGetAllQueueNamesRequest) (*types.QueryGetAllQueueNamesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	names := []string{}

	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, supported := range k.registry.slice {
		queues := whoops.Must(supported.SupportedQueues(ctx))
		for _, q := range queues {
			names = append(names, q.QueueTypeName)
		}
	}

	return &types.QueryGetAllQueueNamesResponse{
		Queues: names,
	}, nil
}
