package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForRelaying(goCtx context.Context, req *types.QueryQueuedMessagesForRelayingRequest) (*types.QueryQueuedMessagesForRelayingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesForRelaying(ctx, req.QueueTypeName, req.ValAddress)
	if err != nil {
		return nil, err
	}

	var res []types.MessageWithSignatures
	for _, msg := range msgs {
		if msg.GetRequireSignatures() {
			msgWithSignatures, err := k.queuedMessageToMessageWithSignatures(msg)
			if err != nil {
				return nil, err
			}
			res = append(res, msgWithSignatures)
		}
	}

	return &types.QueryQueuedMessagesForRelayingResponse{
		Messages: res,
	}, nil
}
