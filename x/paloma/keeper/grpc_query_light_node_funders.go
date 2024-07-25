package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClientFunders(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientFundersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	funders, err := k.LightNodeClientFunders(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientFundersResponse{
		LightNodeClientFunders: funders,
	}, nil
}
