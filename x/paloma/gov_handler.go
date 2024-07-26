package paloma

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/paloma/keeper"
	"github.com/palomachain/paloma/x/paloma/types"
)

func NewPalomaProposalHandler(k keeper.Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.SetLightNodeClientFeegranterProposal:
			acct, err := sdk.AccAddressFromBech32(c.FeegranterAccount)
			if err != nil {
				return err
			}

			return k.SetLightNodeClientFeegranter(ctx, acct)
		case *types.SetLightNodeClientFundersProposal:
			accts := make([]sdk.AccAddress, len(c.FunderAccounts))

			for i, addr := range c.FunderAccounts {
				acct, err := sdk.AccAddressFromBech32(addr)
				if err != nil {
					return err
				}

				accts[i] = acct
			}

			return k.SetLightNodeClientFunders(ctx, accts)
		}

		return sdkerrors.ErrUnknownRequest
	}
}
