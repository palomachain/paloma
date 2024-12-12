package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

func (k Keeper) mintTo(ctx context.Context, amount sdk.Coin, mintTo string) error {
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(mintTo)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
		addr,
		sdk.NewCoins(amount))
}

func (k Keeper) burnFrom(ctx context.Context, amount sdk.Coin, burnFrom string) error {
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(burnFrom)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		addr,
		types.ModuleName,
		sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
}
