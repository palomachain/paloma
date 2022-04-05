package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/volumefi/cronchain/testutil/keeper"
	"github.com/volumefi/cronchain/x/concensus/keeper"
	"github.com/volumefi/cronchain/x/concensus/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.ConcensusKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
