package bech32ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/bech32ibc/keeper"
	"github.com/palomachain/paloma/x/bech32ibc/types"
)

// InitGenesis initializes the module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	err := k.SetNativeHrp(ctx, genState.NativeHRP)
	if err != nil {
		panic(err)
	}
	err = k.SetHrpIbcRecords(ctx, genState.HrpIBCRecords)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	hrpIbcRecords := k.GetHrpIbcRecords(ctx)
	nativeHrp, err := k.GetNativeHrp(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		NativeHRP:     nativeHrp,
		HrpIBCRecords: hrpIbcRecords,
	}
}
