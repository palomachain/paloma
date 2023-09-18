package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
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

func (k msgServer) AddStatusUpdate(goCtx context.Context, msg *types.MsgAddStatusUpdate) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	valAddr := sdk.ValAddress(creator.Bytes())
	status := msg.GetStatus()

	var logFn func(string, ...interface{})
	switch msg.Level {
	case types.MsgAddStatusUpdate_LEVEL_DEBUG:
		logFn = k.Logger(ctx).Debug
	case types.MsgAddStatusUpdate_LEVEL_INFO:
		logFn = k.Logger(ctx).Info
	case types.MsgAddStatusUpdate_LEVEL_ERROR:
		logFn = k.Logger(ctx).Error
	}

	logFn("Pigeon status update",
		"status", status,
		"sender", valAddr)

	return &types.EmptyResponse{}, nil
}
