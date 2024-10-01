package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetValidatorAliveUntil(goCtx context.Context, req *types.QueryGetValidatorAliveUntilRequest) (*types.QueryGetValidatorAliveUntilResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	data, err := k.ValidatorKeepAliveData(ctx, req.GetValAddress())
	if err != nil {
		return nil, err
	}

	return &types.QueryGetValidatorAliveUntilResponse{
		AliveUntilBlockHeight: data.AliveUntilBlockHeight,
	}, nil
}
