package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/volumefi/cronchain/app"
	"github.com/volumefi/cronchain/x/scheduler/keeper"
	"github.com/volumefi/cronchain/x/scheduler/types"
)

var (
	BobAccAddr = sdk.AccAddress("bkla")
)

func setupMsgServer() (types.MsgServer, app.TestApp, context.Context) {
	testApp, ctx := appForTesting()
	return keeper.NewMsgServerImpl(testApp.SchedulerKeeper), testApp, sdk.WrapSDKContext(ctx)
}

func TestRecurringJobSubmission(t *testing.T) {
	type submitRecurringJobTest struct {
		testName                 string
		msgSubmitRecurringJob    *types.MsgSubmitRecurringJob
		msgSubmitRecurringJobRes *types.MsgSubmitRecurringJobResponse
		expectedErr              error
		beforeRunningTest        func(sdk.Context)
		afterRunningTest         func(sdk.Context)
	}

	testData := []submitRecurringJobTest{
		{
			"happy path returns no error",
			types.NewMsgSubmitRecurringJob(
				"abc",
				"bla",
				"bla",
				"blaa",
			),
			&types.MsgSubmitRecurringJobResponse{},
			nil,
			nil,
			nil,
		},
	}

	for _, td := range testData {
		t.Run(td.testName, func(t *testing.T) {
			msgSvr, _, ctx := setupMsgServer()
			sdkCtx := sdk.UnwrapSDKContext(ctx)

			if td.beforeRunningTest != nil {
				td.beforeRunningTest(sdkCtx)
			}

			response, err := msgSvr.SubmitRecurringJob(ctx, td.msgSubmitRecurringJob)

			assert.Equal(t, td.msgSubmitRecurringJobRes, response)
			assert.Equal(t, td.expectedErr, err)

			if td.afterRunningTest != nil {
				td.afterRunningTest(sdkCtx)
			}
		})
	}
}
