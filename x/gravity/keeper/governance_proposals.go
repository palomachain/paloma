package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/gravity/types"
)

func NewGravityProposalHandler(k Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.SetERC20ToDenomProposal:
			ethAddr, err := types.NewEthAddress(c.GetErc20())
			if err != nil {
				return err
			}
			return k.setDenomToERC20(ctx, c.GetChainReferenceId(), c.GetDenom(), *ethAddr)
		case *types.SetBridgeTaxProposal:
			rate, err := strconv.ParseFloat(c.Rate, 32)
			if err != nil {
				return err
			}

			bridgeTax := &types.BridgeTax{
				Rate:            float32(rate),
				ExcludedTokens:  c.ExcludedTokens,
				ExemptAddresses: c.ExemptAddresses,
			}

			return k.SetBridgeTax(ctx, bridgeTax)
		}

		return sdkerrors.ErrUnknownRequest
	}
}
