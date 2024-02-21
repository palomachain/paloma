package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func (k msgServer) CreateJob(goCtx context.Context, msg *types.MsgCreateJob) (*types.MsgCreateJobResponse, error) {
	ctx := common.SdkContext(goCtx)

	var err error
	job := msg.Job
	job.Owner, err = sdk.AccAddressFromBech32(msg.GetCreator())
	if err != nil {
		return nil, err
	}

	if err = k.Keeper.AddNewJob(ctx, job); err != nil {
		return nil, err
	}

	return &types.MsgCreateJobResponse{}, nil
}
