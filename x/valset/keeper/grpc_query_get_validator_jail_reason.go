package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/vizualni/whoops"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetValidatorJailReason(goCtx context.Context, req *types.QueryGetValidatorJailReasonRequest) (*types.QueryGetValidatorJailReasonResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.IsJailed(ctx, req.GetValAddress()) {
		return nil, whoops.String("validator is not jailed")
	}

	reason := string(k.jailReasonStore(ctx).Get([]byte(req.GetValAddress())))
	if reason == "" {
		reason = "validator was offline"
	}

	return &types.QueryGetValidatorJailReasonResponse{
		Reason: reason,
	}, nil
}
