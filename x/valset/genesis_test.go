package valset_test

import (
	"testing"

	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/testutil/nullify"
	"github.com/palomachain/paloma/x/valset"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PigeonRequirements: &types.PigeonRequirements{
			MinVersion: "v10.5.0",
		},
		ScheduledPigeonRequirements: &types.ScheduledPigeonRequirements{
			Requirements: &types.PigeonRequirements{
				MinVersion: "v10.5.0",
			},
			TargetBlockHeight: 1000,
		},
	}

	k, ctx := keepertest.ValsetKeeper(t)
	valset.InitGenesis(ctx, *k, genesisState)
	got := valset.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PigeonRequirements, got.PigeonRequirements)
	require.Equal(t, genesisState.ScheduledPigeonRequirements, got.ScheduledPigeonRequirements)
}
