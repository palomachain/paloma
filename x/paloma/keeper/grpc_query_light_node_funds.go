package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClientFunds(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientFundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	funds, err := k.AllLightNodeClientFunds(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientFundsResponse{
		LightNodeClientFunds: funds,
	}, nil
}
