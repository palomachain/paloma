package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/evm/types"
)

func (k msgServer) SubmitNewJob(goCtx context.Context, msg *types.MsgSubmitNewJob) (*types.MsgSubmitNewJobResponse, error) {
	// TODO: remove this
	panic("do not use this")
}
