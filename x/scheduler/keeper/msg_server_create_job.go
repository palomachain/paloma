package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func (k msgServer) CreateJob(goCtx context.Context, msg *types.MsgCreateJob) (*types.MsgCreateJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var err error
	fmt.Printf("[msgcreatejob_msg] UNPACK ARGS: %+v\n", msg)
	job := msg.Job
	job.Owner, err = sdk.AccAddressFromBech32(msg.GetCreator())
	if err != nil {
		return nil, err
	}

	addr, err := k.AddNewJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateJobResponse{JobAddress: addr}, nil
}
