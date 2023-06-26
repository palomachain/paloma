package treasury_test

import (
	"testing"

	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/testutil/nullify"
	"github.com/palomachain/paloma/x/treasury"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.TreasuryKeeper(t)
	treasury.InitGenesis(ctx, *k, genesisState)
	got := treasury.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	require.Equal(t, "0.01", got.TreasuryFees.SecurityFee)
	require.Equal(t, "0.01", got.TreasuryFees.CommunityFundFee)
	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
