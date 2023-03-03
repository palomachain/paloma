package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/x/valset/keeper"
	"github.com/palomachain/paloma/x/valset/types"
)

//nolint:deadcode
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.ValsetKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

func TestRegisterConductor(t *testing.T) {
	// tapp := app.NewTestApp(false)

	// tapp.MsgServiceRouter
}
