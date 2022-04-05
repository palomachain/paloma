package keeper

import (
	"context"

	"github.com/volumefi/cronchain/x/concensus/types"
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
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// err := k.PutMessageForSigning(ctx, "m", &types.MsgAddMessagesSignatures{
	// 	Creator: "bob",
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// for _, signature := range msg.Signatures {
	// 	if err := AddMessageSignature(ctx, k, signature.MsgID, signature.Signature); err != nil {
	// 		return nil, err
	// 	}
	// }

	return &types.MsgAddMessagesSignaturesResponse{}, nil
}
