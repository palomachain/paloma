package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ChainsInfos(goCtx context.Context, req *types.QueryChainsInfosRequest) (*types.QueryChainsInfosResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := common.SdkContext(goCtx)

	chainsInfos, err := k.GetAllChainInfos(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryChainsInfosResponse{
		ChainsInfos: chainsInfos,
	}, nil
}
