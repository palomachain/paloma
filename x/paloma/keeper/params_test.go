package keeper_test

import (
	"testing"

	testkeeper "github.com/palomachain/paloma/v2/testutil/keeper"
	"github.com/palomachain/paloma/v2/testutil/rand"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.PalomaKeeper(t)
	params := types.Params{
		GasExemptAddresses: []string{
			rand.AccAddress().String(),
		},
	}

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
