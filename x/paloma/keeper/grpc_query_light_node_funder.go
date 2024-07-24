package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClientFunder(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientFunderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	funder, err := k.LightNodeClientFunder(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientFunderResponse{
		LightNodeClientFunder: funder,
	}, nil
}
