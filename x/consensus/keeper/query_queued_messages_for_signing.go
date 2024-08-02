package keeper

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueuedMessagesForSigning(goCtx context.Context, req *types.QueryQueuedMessagesForSigningRequest) (*types.QueryQueuedMessagesForSigningResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesForSigning(ctx, req.QueueTypeName, req.ValAddress)
	if err != nil {
		return nil, err
	}

	var res []*types.MessageToSign
	for _, msg := range msgs {
		if msg.GetRequireSignatures() {
			res = append(res, queuedMessageToMessageToSign(k.cdc, msg))
		}
	}

	return &types.QueryQueuedMessagesForSigningResponse{
		MessageToSign: res,
	}, nil
}

func queuedMessageToMessageToSign(unpacker types.AnyUnpacker, msg types.QueuedSignedMessageI) *types.MessageToSign {
	consensusMsg, err := msg.ConsensusMsg(unpacker)
	if err != nil {
		panic(err)
	}
	anyMsg, err := codectypes.NewAnyWithValue(consensusMsg)
	if err != nil {
		panic(err)
	}
	return &types.MessageToSign{
		Nonce:       nonceFromID(msg.GetId()),
		Id:          msg.GetId(),
		BytesToSign: msg.GetBytesToSign(),
		Msg:         anyMsg,
	}
}
