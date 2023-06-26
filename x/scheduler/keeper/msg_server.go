package keeper

import (
	"github.com/palomachain/paloma/x/scheduler/types"
)

type msgServer struct {
	types.Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
