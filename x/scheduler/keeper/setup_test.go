package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
)

func appForTesting() (app.TestApp, sdk.Context) {
	app := app.NewTestApp(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	return app, ctx
}
func setupMsgServer() (types.MsgServer, app.TestApp, context.Context) {
	testApp, ctx := appForTesting()
	return keeper.NewMsgServerImpl(testApp.SchedulerKeeper), testApp, sdk.WrapSDKContext(ctx)
}
