package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClientFeegranter(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientFeegranterResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	feegranter, err := k.LightNodeClientFeegranter(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientFeegranterResponse{
		LightNodeClientFeegranter: feegranter,
	}, nil
}
