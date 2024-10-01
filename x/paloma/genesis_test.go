package paloma_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/testutil/common"
	keepertest "github.com/palomachain/paloma/v2/testutil/keeper"
	"github.com/palomachain/paloma/v2/testutil/nullify"
	"github.com/palomachain/paloma/v2/x/paloma"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	common.SetupPalomaPrefixes()
	clientAddr := "paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk"

	t.Run("Default params", func(t *testing.T) {
		genesisState := types.GenesisState{
			Params: types.DefaultParams(),
		}

		k, ctx := keepertest.PalomaKeeper(t)
		paloma.InitGenesis(ctx, *k, genesisState)
		got := paloma.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		nullify.Fill(&genesisState)
		nullify.Fill(got)
	})

	t.Run("With light node client licenses", func(t *testing.T) {
		genesisState := types.GenesisState{
			LightNodeClientLicenses: []*types.LightNodeClientLicense{
				{
					ClientAddress: clientAddr,
					Amount: sdk.Coin{
						Amount: math.NewInt(100),
						Denom:  "testgrain",
					},
					VestingMonths: 12,
				},
			},
		}

		k, ctx := keepertest.PalomaKeeper(t)
		paloma.InitGenesis(ctx, *k, genesisState)
		got := paloma.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		require.Equal(t, genesisState.LightNodeClientLicenses,
			got.LightNodeClientLicenses)
	})

	t.Run("With light node client feegranter", func(t *testing.T) {
		accAddr, err := sdk.AccAddressFromBech32(clientAddr)
		require.NoError(t, err)

		genesisState := types.GenesisState{
			LightNodeClientFeegranter: &types.LightNodeClientFeegranter{
				Account: accAddr,
			},
		}

		k, ctx := keepertest.PalomaKeeper(t)
		paloma.InitGenesis(ctx, *k, genesisState)
		got := paloma.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		require.Equal(t, genesisState.LightNodeClientFeegranter.Account,
			got.LightNodeClientFeegranter.Account)
	})

	t.Run("With light node client funder", func(t *testing.T) {
		accAddr, err := sdk.AccAddressFromBech32(clientAddr)
		require.NoError(t, err)

		genesisState := types.GenesisState{
			LightNodeClientFunders: &types.LightNodeClientFunders{
				Accounts: []sdk.AccAddress{accAddr},
			},
		}

		k, ctx := keepertest.PalomaKeeper(t)
		paloma.InitGenesis(ctx, *k, genesisState)
		got := paloma.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		require.Equal(t, genesisState.LightNodeClientFunders.Accounts,
			got.LightNodeClientFunders.Accounts)
	})
}
