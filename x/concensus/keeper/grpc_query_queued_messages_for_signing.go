package keeper

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/concensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForSigning(goCtx context.Context, req *types.QueryQueuedMessagesForSigningRequest) (*types.QueryQueuedMessagesForSigningResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesForSigning(ctx, req.QueueTypeName, sdk.ValAddress(req.ValAddress))
	if err != nil {
		return nil, err
	}
	var res []*codectypes.Any

	for _, msg := range msgs {
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		res = append(res, anyMsg)
	}

	return &types.QueryQueuedMessagesForSigningResponse{
		Msgs: res,
	}, nil
}
