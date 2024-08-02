package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForGasEstimation(goCtx context.Context, req *types.QueryQueuedMessagesForGasEstimationRequest) (*types.QueryQueuedMessagesForGasEstimationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesForGasEstimation(ctx, req.GetQueueTypeName(), req.GetValAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to query messages for gas estimation: %w", err)
	}

	var res []types.MessageToSign
	for _, msg := range msgs {
		res = append(res, *queuedMessageToMessageToSign(k.cdc, msg))
	}

	return &types.QueryQueuedMessagesForGasEstimationResponse{
		MessagesToEstimate: res,
	}, nil
}
