package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func (msgSrv msgServer) ExecuteJob(goCtx context.Context, msg *types.MsgExecuteJob) (*types.MsgExecuteJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := liblog.FromSDKLogger(msgSrv.Keeper.Logger(ctx)).WithFields("job-id", msg.GetJobID())
	logger.Debug("Received ExecuteJob message.")

	// Find the public key of the sender
	creator, err := sdk.AccAddressFromBech32(msg.GetMetadata().GetCreator())
	if err != nil {
		logger.WithError(err).Error("Failed to parse message creator.")
		return nil, err
	}

	senderAddress := msgSrv.Keeper.GetAccount(ctx, creator).GetAddress()

	msgID, err := msgSrv.Keeper.ExecuteJob(ctx, msg.GetJobID(), msg.GetPayload(), senderAddress, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to trigger job execution.")
		return nil, err
	}

	logger.Debug("Job execution triggered.")
	return &types.MsgExecuteJobResponse{
		MessageID: msgID,
	}, nil
}
