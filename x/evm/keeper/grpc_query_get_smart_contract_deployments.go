package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryGetSmartContractDeployments(goCtx context.Context, req *types.QueryGetSmartContractDeploymentsRequest) (*types.QueryGetSmartContractDeploymentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	all, err := k.AllSmartContractsDeployments(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetSmartContractDeploymentsResponse{
		Deployments: all,
	}, nil
}
