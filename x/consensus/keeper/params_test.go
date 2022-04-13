package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/volumefi/cronchain/testutil/keeper"
	"github.com/volumefi/cronchain/x/consensus/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ConsensusKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
