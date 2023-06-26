package treasury

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	err := k.SetCommunityFundFee(ctx, "0.01")
	if err != nil {
		panic(err)
	}

	err = k.SetSecurityFee(ctx, "0.01")
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	fees, err := k.GetFees(ctx)
	if err != nil {
		panic(err)
	}

	genesis.TreasuryFees = fees
	return genesis
}
