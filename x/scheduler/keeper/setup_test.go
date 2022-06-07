package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func appForTesting(t testing.TB) (app.TestApp, sdk.Context) {
	app := app.NewTestApp(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	return app, ctx
}
func setupMsgServer(t testing.TB) (types.MsgServer, app.TestApp, context.Context) {
	testApp, ctx := appForTesting(t)
	return keeper.NewMsgServerImpl(testApp.SchedulerKeeper), testApp, sdk.WrapSDKContext(ctx)
}
