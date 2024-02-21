package keeper_test

import (
	"testing"

	testkeeper "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.SchedulerKeeper(t)
	wctx := common.SdkContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
