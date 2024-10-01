package scheduler_test

import (
	"testing"

	keepertest "github.com/palomachain/paloma/v2/testutil/keeper"
	"github.com/palomachain/paloma/v2/testutil/nullify"
	"github.com/palomachain/paloma/v2/x/scheduler"
	"github.com/palomachain/paloma/v2/x/scheduler/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
	}

	k, ctx := keepertest.SchedulerKeeper(t)
	scheduler.InitGenesis(ctx, *k, genesisState)
	got := scheduler.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)
}
