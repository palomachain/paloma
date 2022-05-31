package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetAllQueueNames(goCtx context.Context, req *types.QueryGetAllQueueNamesRequest) (*types.QueryGetAllQueueNamesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	names := []string{}

	for queueTypeName := range k.queueRegistry {
		names = append(names, string(queueTypeName))
	}

	return &types.QueryGetAllQueueNamesResponse{
		Queues: names,
	}, nil
}
