package keeper

import (
	"github.com/palomachain/paloma/x/scheduler/types"
)

type msgServer struct {
	SchedulerKeeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{SchedulerKeeper: keeper}
}

var _ types.MsgServer = msgServer{}
