package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func (k msgServer) ExecuteJob(goCtx context.Context, msg *types.MsgExecuteJob) (*types.MsgExecuteJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Find the public key of the sender
	pubKeyBytes := k.account.GetAccount(ctx, msg.GetSigners()[0]).GetPubKey().Bytes()

	err := k.Keeper.ScheduleNow(ctx, msg.GetJobID(), msg.GetPayload(), pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteJobResponse{}, nil
}
