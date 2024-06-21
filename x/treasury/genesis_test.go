package treasury_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/palomachain/paloma/testutil/keeper"
	"github.com/palomachain/paloma/testutil/nullify"
	"github.com/palomachain/paloma/x/treasury"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	relayerFees := []types.RelayerFeeSetting{
		{
			ValAddress: sdk.ValAddress("validator-1").String(),
			Fees: []types.RelayerFeeSetting_FeeSetting{
				{
					Multiplicator:    math.LegacyMustNewDecFromStr("0.5"),
					ChainReferenceId: "test-chain",
				},
			},
		},
		{
			ValAddress: sdk.ValAddress("validator-2").String(),
			Fees: []types.RelayerFeeSetting_FeeSetting{
				{
					Multiplicator:    math.LegacyMustNewDecFromStr("2.5"),
					ChainReferenceId: "test-chain",
				},
			},
		},
	}
	genesisState := types.GenesisState{
		Params: types.Params{},
		TreasuryFees: types.Fees{
			CommunityFundFee: "2.0",
			SecurityFee:      "0.5",
		},
		RelayerFeeSettings: relayerFees,
	}

	k, ctx := keepertest.TreasuryKeeper(t)
	treasury.InitGenesis(ctx, *k, genesisState)
	got := treasury.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	require.Equal(t, "0.5", got.TreasuryFees.SecurityFee)
	require.Equal(t, "2.0", got.TreasuryFees.CommunityFundFee)
	require.Len(t, got.RelayerFeeSettings, 2)
	require.Equal(t, relayerFees, got.RelayerFeeSettings)
	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
