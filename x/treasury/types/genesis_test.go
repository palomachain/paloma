package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/treasury/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{},
				TreasuryFees: types.Fees{
					CommunityFundFee: "2.0",
					SecurityFee:      "0.5",
				},
				RelayerFeeSettings: []types.RelayerFeeSetting{
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
				},
			},
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
