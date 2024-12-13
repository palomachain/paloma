package types_test

import (
	"testing"

	testutilcommmon "github.com/palomachain/paloma/v2/testutil/common"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	testutilcommmon.SetupPalomaPrefixes()
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
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "different admin from creator",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "paloma1kludne80z0tcq9t7j9fqa630fechsjhxhpafac",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "empty admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "no admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
					},
				},
			},
			valid: true,
		},
		{
			desc: "invalid admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "moose",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "multiple denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/litecoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicate denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: false,
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
