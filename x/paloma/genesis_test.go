package paloma_test

import (
	"testing"

	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/testutil/nullify"
	"github.com/palomachain/paloma/x/paloma"
	"github.com/palomachain/paloma/x/paloma/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.PalomaKeeper(t)
	paloma.InitGenesis(ctx, *k, genesisState)
	got := paloma.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
