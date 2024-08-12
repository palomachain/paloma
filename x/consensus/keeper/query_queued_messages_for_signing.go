package keeper

import (
	"context"
	"encoding/hex"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
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
			res = append(res, k.queuedMessageToMessageToSign(ctx, msg))
		}
	}

	return &types.QueryQueuedMessagesForSigningResponse{
		MessageToSign: res,
	}, nil
}

func (k Keeper) queuedMessageToMessageToSign(ctx context.Context, msg types.QueuedSignedMessageI) *types.MessageToSign {
	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		panic(err)
	}

	anyMsg, err := codectypes.NewAnyWithValue(consensusMsg)
	if err != nil {
		panic(err)
	}

	bytesToSign, err := msg.GetBytesToSign2(k.cdc)
	if err != nil {
		panic(err)
	}

	liblog.FromKeeper(ctx, k).
		WithFields(
			"testing", "TESTING",
			"bytes_to_sign", hex.EncodeToString(msg.GetBytesToSign()),
			"new_bytes_to_sign", hex.EncodeToString(bytesToSign),
		).Debug("New bytes to sign calculation")

	return &types.MessageToSign{
		Nonce:       nonceFromID(msg.GetId()),
		Id:          msg.GetId(),
		BytesToSign: bytesToSign,
		Msg:         anyMsg,
	}
}
