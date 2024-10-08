package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClientActivations(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientActivationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	activations, err := k.AllLightNodeClientActivations(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientActivationsResponse{
		LightNodeClientActivations: activations,
	}, nil
}
