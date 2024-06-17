package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

func (k Keeper) GetBridgeTax(
	goCtx context.Context,
	req *types.QueryBridgeTaxRequest,
) (*types.QueryBridgeTaxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tax, err := k.BridgeTax(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryBridgeTaxResponse{
		BridgeTax: tax,
	}, nil
}
