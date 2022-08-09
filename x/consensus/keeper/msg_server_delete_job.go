package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/consensus/types"
)

func (k msgServer) DeleteJob(goCtx context.Context, msg *types.MsgDeleteJob) (*types.MsgDeleteJobResponse, error) {
	panic("do not use this!")
}
