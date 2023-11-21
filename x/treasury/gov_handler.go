package treasury

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/treasury/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

func NewFeeProposalHandler(k keeper.Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.CommunityFundFeeProposal:
			return k.SetCommunityFundFee(ctx, c.GetFee())
		case *types.SecurityFeeProposal:
			return k.SetSecurityFee(ctx, c.GetFee())
		}

		return sdkerrors.ErrUnknownRequest
	}
}
