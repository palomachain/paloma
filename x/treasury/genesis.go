package treasury

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	sdkctx := sdk.UnwrapSDKContext(ctx)
	k.SetParams(sdkctx, genState.Params)

	err := k.SetCommunityFundFee(sdkctx, genState.TreasuryFees.CommunityFundFee)
	if err != nil {
		panic(err)
	}

	err = k.SetSecurityFee(sdkctx, genState.TreasuryFees.SecurityFee)
	if err != nil {
		panic(err)
	}

	for _, v := range genState.RelayerFeeSettings {
		addr, err := sdk.ValAddressFromBech32(v.ValAddress)
		if err != nil {
			panic(fmt.Errorf("set relayer fee: parse val addr: %w", err))
		}
		if err := k.SetRelayerFee(ctx, addr, &v); err != nil {
			panic(fmt.Errorf("set relayer fee: %w", err))
		}
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

	genesis.TreasuryFees = *fees

	relayerFees, err := k.GetRelayerFees(ctx)
	if err != nil {
		panic(err)
	}

	genesis.RelayerFeeSettings = relayerFees
	return genesis
}
