package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetValidatorJailReason(goCtx context.Context, req *types.QueryGetValidatorJailReasonRequest) (*types.QueryGetValidatorJailReasonResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.IsJailed(ctx, req.GetValAddr()) {
		return k.
	}

	reason := k.jailReasonStore(ctx).Get([]byte(req.GetValAddr()))
	if reason == "" {

	}

	return &types.QueryGetValidatorJailReasonResponse{}, nil
}
