package paloma

import (
	"context"
	"errors"

	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/paloma/keeper"
	"github.com/palomachain/paloma/x/paloma/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, license := range genState.LightNodeClientLicenses {
		err := k.SetLightNodeClientLicense(ctx, license.ClientAddress, license)
		if err != nil {
			panic(err)
		}
	}

	if genState.LightNodeClientFeegranter != nil {
		err := k.SetLightNodeClientFeegranter(ctx,
			genState.LightNodeClientFeegranter.Account)
		if err != nil {
			panic(err)
		}
	}

	if genState.LightNodeClientFunders != nil {
		err := k.SetLightNodeClientFunders(ctx,
			genState.LightNodeClientFunders.Accounts)
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

	genesis.LightNodeClientLicenses, err = k.AllLightNodeClientLicenses(ctx)
	if err != nil {
		panic(err)
	}

	genesis.LightNodeClientFeegranter, err = k.LightNodeClientFeegranter(ctx)
	if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
		panic(err)
	}

	genesis.LightNodeClientFunders, err = k.LightNodeClientFunders(ctx)
	if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
		panic(err)
	}

	return genesis
}
