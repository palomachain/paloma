package valset

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/x/valset/keeper"
	"github.com/palomachain/paloma/v2/x/valset/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	err := k.SetPigeonRequirements(ctx, genState.PigeonRequirements)
	if err != nil {
		panic(err)
	}

	err = k.SetScheduledPigeonRequirements(ctx, genState.ScheduledPigeonRequirements)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	pigeonRequirements, err := k.PigeonRequirements(ctx)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			panic(err)
		}
	} else {
		genesis.PigeonRequirements = pigeonRequirements
	}

	scheduledPigeonRequirements, err := k.ScheduledPigeonRequirements(ctx)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			panic(err)
		}
	} else {
		genesis.ScheduledPigeonRequirements = scheduledPigeonRequirements
	}

	return genesis
}
