package keeper

import (
	"context"

    "github.com/palomachain/paloma/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ChainsInfos(goCtx context.Context,  req *types.QueryChainsInfosRequest) (*types.QueryChainsInfosResponse, error) {
	if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Process the query
    _ = ctx

	return &types.QueryChainsInfosResponse{}, nil
}
