package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetAlivePigeons(goCtx context.Context, req *types.QueryGetAlivePigeonsRequest) (*types.QueryGetAlivePigeonsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	vals := k.unjailedValidators(ctx)

	now := time.Now().UTC()
	res := slice.Map(vals, func(val stakingtypes.ValidatorI) *types.QueryGetAlivePigeonsResponse_ValidatorAlive {
		until, err := k.ValidatorAliveUntil(ctx, val.GetOperator())
		s := &types.QueryGetAlivePigeonsResponse_ValidatorAlive{
			ValAddress: val.GetOperator(),
		}
		if err != nil {
			s.Error = err.Error()
		} else {
			s.AliveUntil = until
			s.Ttl = int64(until.Sub(now))
		}
		return s
	})

	return &types.QueryGetAlivePigeonsResponse{
		AliveValidators: res,
	}, nil
}
