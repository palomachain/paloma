package keeper

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestIBCMetadataProposal(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	ibcDenom := "ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED/grav"
	goodProposal := types.IBCMetadataProposal{
		Title:       "test tile",
		Description: "test description",
		Metadata: banktypes.Metadata{
			Description: "Atom",
			Name:        "Atom",
			Base:        ibcDenom,
			Display:     "Atom",
			Symbol:      "ATOM",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    ibcDenom,
					Exponent: 0,
				},
				{
					Denom:    "Atom",
					Exponent: 6,
				},
			},
		},
		IbcDenom: ibcDenom,
	}

	gk := input.GravityKeeper

	err := gk.HandleIBCMetadataProposal(ctx, &goodProposal)
	require.NoError(t, err)
	metadata, exists := gk.bankKeeper.GetDenomMetaData(ctx, ibcDenom)
	require.True(t, exists)
	require.Equal(t, metadata, goodProposal.Metadata)

	// does not have a zero base unit
	badMetadata := goodProposal
	badMetadata.Metadata.DenomUnits = []*banktypes.DenomUnit{
		{
			Denom:    ibcDenom,
			Exponent: 1,
		},
		{
			Denom:    "Atom",
			Exponent: 6,
		},
	}

	err = gk.HandleIBCMetadataProposal(ctx, &badMetadata)
	require.Error(t, err)

	// no denom unit for display
	badMetadata2 := goodProposal
	badMetadata2.Metadata.DenomUnits = []*banktypes.DenomUnit{
		{
			Denom:    ibcDenom,
			Exponent: 0,
		},
		{
			Denom:    "a",
			Exponent: 6,
		},
	}

	err = gk.HandleIBCMetadataProposal(ctx, &badMetadata2)
	require.Error(t, err)

	// incorrect base unit
	badMetadata3 := goodProposal
	badMetadata3.Metadata.DenomUnits = []*banktypes.DenomUnit{
		{
			Denom:    "atom",
			Exponent: 0,
		},
		{
			Denom:    "a",
			Exponent: 6,
		},
	}

	err = gk.HandleIBCMetadataProposal(ctx, &badMetadata3)
	require.Error(t, err)

}
