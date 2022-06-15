package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetSnapshotByID(goCtx context.Context, req *types.QueryGetSnapshotByIDRequest) (*types.QueryGetSnapshotByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	snapshot, err := k.getSnapshotByID(ctx, req.SnapshotId)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetSnapshotByIDResponse{
		Snapshot: snapshot,
	}, nil
}
