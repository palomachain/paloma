package keeper_test

import (
	"testing"

	testkeeper "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/x/paloma/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.PalomaKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
