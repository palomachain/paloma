package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/volumefi/cronchain/app"
)

func appForTesting() (app.TestApp, sdk.Context) {
	app := app.NewTestApp(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	return app, ctx
}
