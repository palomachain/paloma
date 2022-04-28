package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func nonceFromID(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
}

func (k Keeper) QueuedMessagesForSigning(goCtx context.Context, req *types.QueryQueuedMessagesForSigningRequest) (*types.QueryQueuedMessagesForSigningResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesForSigning(ctx, req.QueueTypeName, sdk.ValAddress(req.ValAddress))
	if err != nil {
		return nil, err
	}

	var res []*types.MessageToSign
	for _, msg := range msgs {
		res = append(res, &types.MessageToSign{
			Nonce: nonceFromID(msg.GetId()),
			Id:    msg.GetId(),
			Msg:   msg.GetMsg(),
		})
	}

	return &types.QueryQueuedMessagesForSigningResponse{
		MessageToSign: res,
	}, nil
}
