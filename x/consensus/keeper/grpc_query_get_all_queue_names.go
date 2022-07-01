package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/vizualni/whoops"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetAllQueueNames(goCtx context.Context, req *types.QueryGetAllQueueNamesRequest) (*types.QueryGetAllQueueNamesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	names := []string{}

	for _, supported := range k.registry.slice {
		for queue := range whoops.Must(supported.SupportedQueues(sdk.UnwrapSDKContext(goCtx))) {
			names = append(names, queue)
		}
	}

	return &types.QueryGetAllQueueNamesResponse{
		Queues: names,
	}, nil
}
