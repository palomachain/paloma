package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ConsensusReached(goCtx context.Context, req *types.QueryConsensusReachedRequest) (*types.QueryConsensusReachedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	snapshot, err := k.valset.GetCurrentSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	cq, err := k.getConsensusQueue(req.QueueTypeName)

	if err != nil {
		return nil, err
	}

	cq.

	return &types.QueryConsensusReachedResponse{}, nil
}
