package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/x/consensus/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.ConsensusKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
