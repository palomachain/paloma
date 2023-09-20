package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetLatestPublishedSnapshot(goCtx context.Context, req *types.QueryGetLatestPublishedSnapshotRequest) (*types.QueryGetLatestPublishedSnapshotResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	snapshot, err := k.GetLatestSnapshotOnChain(ctx, req.ChainReferenceID)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetLatestPublishedSnapshotResponse{
		Snapshot: snapshot,
	}, nil
}
