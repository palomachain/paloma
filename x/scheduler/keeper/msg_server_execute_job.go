package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func (msgSrv msgServer) ExecuteJob(goCtx context.Context, msg *types.MsgExecuteJob) (*types.MsgExecuteJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := liblog.FromSDKLogger(msgSrv.Logger(ctx)).WithFields("job-id", msg.GetJobID())
	logger.Debug("Received ExecuteJob message.")

	// Find the public key of the sender
	senderAddress := msgSrv.GetAccount(ctx, msg.GetSigners()[0]).GetAddress()

	err := msgSrv.Keeper.ExecuteJob(ctx, msg.GetJobID(), msg.GetPayload(), senderAddress, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to trigger job execution.")
		return nil, err
	}

	logger.Debug("Job execution triggered.")
	return &types.MsgExecuteJobResponse{}, nil
}
