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

	job, err := k.GetJob(ctx, msg.GetJobID())
	if err != nil {
		k.Logger(ctx).Error("couldn't get job's id",
			"err", err,
			"job_id", msg.GetJobID(),
		)
		return &types.MsgExecuteJobResponse{}, err
	}

	// Hook to trigger a valset update attempt
	err = k.Keeper.EvmKeeper.OnJobExecution(ctx, job)
	if err != nil {
		// If we have an error here, don't exit.  Go ahead and schedule the job
		k.Logger(ctx).Error("Error in EvmKeeper OnJobExecution hook",
			"err", err,
			"job_id", msg.GetJobID(),
		)
	}

	err = k.Keeper.ScheduleNow(ctx, msg.GetJobID(), msg.GetPayload(), pubKeyBytes, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteJobResponse{}, nil
}
