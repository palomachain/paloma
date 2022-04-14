package valset_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/volumefi/cronchain/testutil/keeper"
	"github.com/volumefi/cronchain/testutil/nullify"
	"github.com/volumefi/cronchain/x/valset"
	"github.com/volumefi/cronchain/x/valset/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ValsetKeeper(t)
	valset.InitGenesis(ctx, *k, genesisState)
	got := valset.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
