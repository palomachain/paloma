package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func (msgSrv msgServer) ExecuteJob(goCtx context.Context, msg *types.MsgExecuteJob) (*types.MsgExecuteJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Find the public key of the sender
	pubKeyBytes := msgSrv.GetAccount(ctx, msg.GetSigners()[0]).GetPubKey().Bytes()

	job, err := msgSrv.GetJob(ctx, msg.GetJobID())
	if err != nil {
		msgSrv.Logger(ctx).Error("couldn't get job's id",
			"err", err,
			"job_id", msg.GetJobID(),
		)
		return nil, err
	}

	// Hook to trigger a valset update attempt
	err = msgSrv.PreJobExecution(ctx, job)
	if err != nil {
		// If we have an error here, don't exit.  Go ahead and schedule the job
		msgSrv.Logger(ctx).Error("Error in PreJobExecution hook",
			"err", err,
			"job_id", msg.GetJobID(),
		)
	}

	err = msgSrv.ScheduleNow(ctx, msg.GetJobID(), msg.GetPayload(), pubKeyBytes, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteJobResponse{}, nil
}
