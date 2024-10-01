package keeper

import (
	"context"

	"github.com/palomachain/paloma/v2/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryUserSmartContracts(
	ctx context.Context,
	req *types.QueryUserSmartContractsRequest,
) (*types.QueryUserSmartContractsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	contracts, err := k.UserSmartContracts(ctx, req.ValAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryUserSmartContractsResponse{
		Contracts: contracts,
	}, nil
}
