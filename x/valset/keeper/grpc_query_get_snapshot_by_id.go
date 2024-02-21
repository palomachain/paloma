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
	var snapshot *types.Snapshot
	var err error

	if req.SnapshotId == 0 {
		snapshot, err = k.GetCurrentSnapshot(ctx)
	} else {
		snapshot, err = k.FindSnapshotByID(ctx, req.SnapshotId)
	}

	if err != nil {
		return nil, err
	}

	return &types.QueryGetSnapshotByIDResponse{
		Snapshot: snapshot,
	}, nil
}
