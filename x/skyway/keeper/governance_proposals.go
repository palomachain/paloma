package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/skyway/types"
)

func NewSkywayProposalHandler(k Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.SetERC20ToDenomProposal:
			ethAddr, err := types.NewEthAddress(c.GetErc20())
			if err != nil {
				return err
			}
			return k.setDenomToERC20(ctx, c.GetChainReferenceId(), c.GetDenom(), *ethAddr)
		case *types.SetBridgeTaxProposal:
			addresses := make([]sdk.AccAddress, 0, len(c.ExemptAddresses))
			for _, addr := range c.ExemptAddresses {
				address, err := sdk.AccAddressFromBech32(addr)
				if err != nil {
					return err
				}

				addresses = append(addresses, address)
			}

			bridgeTax := &types.BridgeTax{
				Rate:            c.Rate,
				Token:           c.Token,
				ExemptAddresses: addresses,
			}

			return k.SetBridgeTax(ctx, bridgeTax)
		case *types.SetBridgeTransferLimitProposal:
			addresses := make([]sdk.AccAddress, 0, len(c.ExemptAddresses))
			for _, addr := range c.ExemptAddresses {
				address, err := sdk.AccAddressFromBech32(addr)
				if err != nil {
					return err
				}

				addresses = append(addresses, address)
			}

			limit := &types.BridgeTransferLimit{
				Token:           c.Token,
				Limit:           c.Limit,
				LimitPeriod:     c.LimitPeriod,
				ExemptAddresses: addresses,
			}

			return k.SetBridgeTransferLimit(ctx, limit)
		}

		return sdkerrors.ErrUnknownRequest
	}
}
