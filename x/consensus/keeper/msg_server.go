package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (k msgServer) AddMessageEstimates(goCtx context.Context, msg *types.MsgAddMessageGasEstimates) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, _ := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	valAddr := sdk.ValAddress(creator.Bytes())

	if err := k.Keeper.valset.CanAcceptValidator(ctx, valAddr); err != nil {
		return nil, err
	}

	if msg.Estimates == nil {
		return nil, nil
	}

	for _, estimate := range msg.Estimates {
		if estimate.GetValue() < 1 {
			return nil, fmt.Errorf("ivnalid gas estimate")
		}
	}

	if err := k.AddMessageGasEstimates(
		ctx,
		valAddr,
		msg.Estimates,
	); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
