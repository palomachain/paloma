package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForSigning(goCtx context.Context, req *types.QueryQueuedMessagesForSigningRequest) (*types.QueryQueuedMessagesForSigningResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := common.SdkContext(goCtx)

	msgs, err := k.GetMessagesForSigning(ctx, req.QueueTypeName, req.ValAddress)
	if err != nil {
		return nil, err
	}

	var res []*types.MessageToSign
	for _, msg := range msgs {
		if msg.GetRequireSignatures() {
			res = append(res, k.queuedMessageToMessageToSign(msg))
		}
	}

	return &types.QueryQueuedMessagesForSigningResponse{
		MessageToSign: res,
	}, nil
}
