package keeper_test

import (
	"testing"

	testkeeper "github.com/palomachain/paloma/v2/testutil/keeper"
	"github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.ConsensusKeeper(t)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
