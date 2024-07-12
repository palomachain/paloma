package paloma

import (
	"context"

	"github.com/palomachain/paloma/x/paloma/keeper"
	"github.com/palomachain/paloma/x/paloma/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, funds := range genState.LightNodeClientFunds {
		err := k.SetLightNodeClientFunds(ctx, funds.ClientAddress, funds)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.LightNodeClientFunds, err = k.AllLightNodeClientFunds(ctx)
	if err != nil {
		panic(err)
	}

	return genesis
}
