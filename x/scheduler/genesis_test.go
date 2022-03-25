package scheduler_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/volumefi/cronchain/testutil/keeper"
	"github.com/volumefi/cronchain/testutil/nullify"
	"github.com/volumefi/cronchain/x/scheduler"
	"github.com/volumefi/cronchain/x/scheduler/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SchedulerKeeper(t)
	scheduler.InitGenesis(ctx, *k, genesisState)
	got := scheduler.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}
