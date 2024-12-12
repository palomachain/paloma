package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) {
	k.CreateModuleAccount(ctx)

	if genState.Params.DenomCreationFee == nil {
		genState.Params.DenomCreationFee = sdk.NewCoins()
	}
	k.SetParams(ctx, genState.Params)

	for _, genDenom := range genState.GetFactoryDenoms() {
		creator, _, err := types.DeconstructDenom(genDenom.GetDenom())
		if err != nil {
			panic(err)
		}
		err = k.createDenomAfterValidation(ctx, creator, genDenom.GetDenom())
		if err != nil {
			panic(err)
		}
		err = k.setAuthorityMetadata(ctx, genDenom.GetDenom(), genDenom.GetAuthorityMetadata())
		if err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	genDenoms := []types.GenesisDenom{}
	iterator := k.GetAllDenomsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Value())

		authorityMetadata, err := k.GetAuthorityMetadata(ctx, denom)
		if err != nil {
			panic(err)
		}

		genDenoms = append(genDenoms, types.GenesisDenom{
			Denom:             denom,
			AuthorityMetadata: authorityMetadata,
		})
	}

	return &types.GenesisState{
		FactoryDenoms: genDenoms,
		Params:        k.GetParams(ctx),
	}
}
