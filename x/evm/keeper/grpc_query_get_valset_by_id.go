package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetValsetByID(goCtx context.Context, req *types.QueryGetValsetByIDRequest) (*types.QueryGetValsetByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	snapshot, err := k.Valset.FindSnapshotByID(ctx, req.GetValsettID())
	if err != nil {
		return nil, err
	}
	valset := transformSnapshotToTurnstoneValset(snapshot, req.GetChainID())

	return &types.QueryGetValsetByIDResponse{
		Valset: &valset,
	}, nil
}
