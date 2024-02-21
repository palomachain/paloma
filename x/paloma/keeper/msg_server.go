package keeper

import (
	"context"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/paloma/types"
)

const cPigeonStatusUpdateFF = "PALOMA_FF_PIGEON_STATUS_UPDATE"

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

	// Avoid log spamming
	_, ok := os.LookupEnv(cPigeonStatusUpdateFF)
	if !ok {
		return &types.EmptyResponse{}, nil
	}

	creator, err := sdk.AccAddressFromBech32(msg.GetMetadata().GetCreator())
	if err != nil {
		err = fmt.Errorf("failed to parse creator address: %w", err)
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithFields("component", "pigeon-status-update").Error("Failed to parse creator from message")
		return nil, err
	}
	valAddr := sdk.ValAddress(creator.Bytes())
	status := msg.GetStatus()

	var logFn func(string, ...interface{})
	switch msg.Level {
	case types.MsgAddStatusUpdate_LEVEL_DEBUG:
		logFn = liblog.FromSDKLogger(k.Logger(ctx)).Debug
	case types.MsgAddStatusUpdate_LEVEL_INFO:
		logFn = liblog.FromSDKLogger(k.Logger(ctx)).Info
	case types.MsgAddStatusUpdate_LEVEL_ERROR:
		logFn = liblog.FromSDKLogger(k.Logger(ctx)).Error
	}

	args := make([]interface{}, 0, len(msg.GetArgs())*2+6) // len: (k+v) + 6 for static appends
	for _, v := range msg.GetArgs() {
		args = append(args, fmt.Sprintf("msg.args.%s", v.GetKey()), v.GetValue())
	}

	args = append(args,
		"component", "pigeon-status-update",
		"status", status,
		"sender", valAddr)

	logFn(status, args...)

	return &types.EmptyResponse{}, nil
}
