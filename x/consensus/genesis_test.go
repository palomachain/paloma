package consensus_test

import (
	"testing"

	keepertest "github.com/palomachain/paloma/v2/testutil/keeper"
	"github.com/palomachain/paloma/v2/testutil/nullify"
	"github.com/palomachain/paloma/v2/x/consensus"
	"github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
	}

	k, ctx := keepertest.ConsensusKeeper(t)
	consensus.InitGenesis(ctx, *k, genesisState)
	got := consensus.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)
}
