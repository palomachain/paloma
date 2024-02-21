package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForRelaying(goCtx context.Context, req *types.QueryQueuedMessagesForRelayingRequest) (*types.QueryQueuedMessagesForRelayingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := liblog.FromSDKLogger(k.Logger(ctx)).WithFields(
		"component", "queued-messages-for-relaying",
		"sender", req.ValAddress.String())

	msgs, err := k.GetMessagesForRelaying(ctx, req.QueueTypeName, req.ValAddress)
	if err != nil {
		logger.WithError(err).Error("Failed to query messages for relaying.")
		return nil, err
	}

	var res []types.MessageWithSignatures
	for _, msg := range msgs {
		if msg.GetRequireSignatures() {
			msgWithSignatures, err := k.queuedMessageToMessageWithSignatures(msg)
			if err != nil {
				logger.WithError(err).Error("Failed to parse queued message to message with signatures.")
				return nil, err
			}
			res = append(res, msgWithSignatures)
		}
	}

	msgIDs := make([]uint64, len(res))
	for i, v := range res {
		msgIDs[i] = v.Id
	}
	logger.WithFields("messages", msgIDs).Debug("Queried for pending messages.")
	return &types.QueryQueuedMessagesForRelayingResponse{
		Messages: res,
	}, nil
}
