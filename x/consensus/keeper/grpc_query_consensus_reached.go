package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConsensusReached returns messages that given a queueTypeName have reched a consensus.
func (k Keeper) ConsensusReached(goCtx context.Context, req *types.QueryConsensusReachedRequest) (*types.QueryConsensusReachedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesThatHaveReachedConsensus(ctx, req.QueueTypeName)

	if err != nil {
		return nil, err
	}

	res := &types.QueryConsensusReachedResponse{
		QueueTypeName: req.QueueTypeName,
	}

	for _, msg := range msgs {
		res.Messages = append(res.Messages, queuedMessageToMessageToSign(msg))
	}

	return &types.QueryConsensusReachedResponse{}, nil
}
