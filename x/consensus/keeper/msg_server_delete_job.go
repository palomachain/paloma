package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

func (k msgServer) DeleteJob(goCtx context.Context, msg *types.MsgDeleteJob) (*types.MsgDeleteJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.deleteJob(ctx, types.ConsensusQueueType(msg.GetQueueTypeName()), msg.GetMessageID()); err != nil {
		return nil, err
	}
	return &types.MsgDeleteJobResponse{}, nil
}
