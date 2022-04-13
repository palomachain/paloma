package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/consensus/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) AddMessagesSignatures(goCtx context.Context, msg *types.MsgAddMessagesSignatures) (*types.MsgAddMessagesSignaturesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: once we implement keeper.AddMessageSignature write te

	for _, signedMessage := range msg.SignedMessages {
		if err := k.AddMessageSignature(
			ctx,
			signedMessage.GetId(),
			signedMessage.GetQueueTypeName(),
			signedMessage.GetSignature(),
		); err != nil {
			return nil, err
		}
	}

	return &types.MsgAddMessagesSignaturesResponse{}, nil
}
