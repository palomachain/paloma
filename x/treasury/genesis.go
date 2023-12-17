package treasury

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	sdkctx := sdk.UnwrapSDKContext(ctx)
	k.SetParams(sdkctx, genState.Params)

	err := k.SetCommunityFundFee(sdkctx, "0.01")
	if err != nil {
		panic(err)
	}

	err = k.SetSecurityFee(sdkctx, "0.01")
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	sdkctx := sdk.UnwrapSDKContext(ctx)

	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(sdkctx)

	fees, err := k.GetFees(sdkctx)
	if err != nil {
		panic(err)
	}

	genesis.TreasuryFees = fees
	return genesis
}
