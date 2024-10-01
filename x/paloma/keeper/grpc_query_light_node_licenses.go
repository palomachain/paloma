package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClientLicenses(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientLicensesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	licenses, err := k.AllLightNodeClientLicenses(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientLicensesResponse{
		LightNodeClientLicenses: licenses,
	}, nil
}
