package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForAttesting(goCtx context.Context, req *types.QueryQueuedMessagesForAttestingRequest) (*types.QueryQueuedMessagesForAttestingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesForAttesting(ctx, req.QueueTypeName, req.ValAddress)
	if err != nil {
		return nil, err
	}

	var res []*types.MessageWithSignatures
	for _, msg := range msgs {
		res = append(res, k.queuedMessageToMessageWithSignatures(msg))
	}

	return &types.QueryQueuedMessagesForAttestingResponse{
		Messages: res,
	}, nil
}
