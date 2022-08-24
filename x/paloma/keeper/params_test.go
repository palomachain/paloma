package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/x/paloma/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.PalomaKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
