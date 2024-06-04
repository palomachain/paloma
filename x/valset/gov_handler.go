package valset

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/valset/keeper"
	"github.com/palomachain/paloma/x/valset/types"
)

func NewValsetProposalHandler(k keeper.Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.SetPigeonRequirementsProposal:
			targetBlockHeight := c.GetTargetBlockHeight()

			if ctx.BlockHeight() >= int64(targetBlockHeight) {
				// If the we're already past the target block height, we apply
				// the requirements immediately. This is also true if the
				// proposal does not contain a target block height.
				return k.SetPigeonRequirements(ctx, &types.PigeonRequirements{
					MinVersion: c.GetMinVersion(),
				})
			}

			return k.SetScheduledPigeonRequirements(ctx, &types.ScheduledPigeonRequirements{
				Requirements: &types.PigeonRequirements{
					MinVersion: c.GetMinVersion(),
				},
				TargetBlockHeight: targetBlockHeight,
			})
		}

		return sdkerrors.ErrUnknownRequest
	}
}
