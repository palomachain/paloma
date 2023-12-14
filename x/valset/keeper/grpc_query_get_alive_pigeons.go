package keeper

import (
	"context"

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

	vals := k.GetUnjailedValidators(ctx)

	res := slice.Map(vals, func(val stakingtypes.ValidatorI) *types.QueryGetAlivePigeonsResponse_ValidatorAlive {
		bz, err := k.addressCodec.StringToBytes(val.GetOperator())
		s := &types.QueryGetAlivePigeonsResponse_ValidatorAlive{
			ValAddress: bz,
		}
		if err != nil {
			s.Error = err.Error()
		}
		until, err := k.ValidatorAliveUntil(ctx, bz)
		if err != nil {
			s.Error = err.Error()
		} else {
			s.AliveUntilBlockHeight = until
		}
		return s
	})

	return &types.QueryGetAlivePigeonsResponse{
		AliveValidators: res,
	}, nil
}
