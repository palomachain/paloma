package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
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

	creator, _ := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	valAddr := sdk.ValAddress(creator.Bytes())

	if err := k.Keeper.valset.CanAcceptValidator(ctx, valAddr); err != nil {
		return nil, err
	}

	if err := k.AddMessageSignature(
		ctx,
		valAddr,
		msg.SignedMessages,
	); err != nil {
		return nil, err
	}

	return &types.MsgAddMessagesSignaturesResponse{}, nil
}
