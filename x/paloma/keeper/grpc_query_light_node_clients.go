package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeClients(
	c context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeClientsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	clients, err := k.AllLightNodeClients(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeClientsResponse{
		LightNodeClients: clients,
	}, nil
}
