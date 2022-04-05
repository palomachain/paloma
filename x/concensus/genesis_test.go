package concensus_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/volumefi/cronchain/testutil/keeper"
	"github.com/volumefi/cronchain/testutil/nullify"
	"github.com/volumefi/cronchain/x/concensus"
	"github.com/volumefi/cronchain/x/concensus/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ConcensusKeeper(t)
	concensus.InitGenesis(ctx, *k, genesisState)
	got := concensus.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}
