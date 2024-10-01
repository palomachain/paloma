package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/skyway/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetBridgeTaxes(
	goCtx context.Context,
	_ *emptypb.Empty,
) (*types.QueryBridgeTaxesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	taxes, err := k.AllBridgeTaxes(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryBridgeTaxesResponse{
		BridgeTaxes: taxes,
	}, nil
}
