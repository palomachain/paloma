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

	if err := k.AddMessageSignature(
		ctx,
		sdk.ValAddress(msg.Creator),
		msg.SignedMessages,
	); err != nil {
		return nil, err
	}

	return &types.MsgAddMessagesSignaturesResponse{}, nil
}
