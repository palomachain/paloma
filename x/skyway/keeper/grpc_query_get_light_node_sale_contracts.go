package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/skyway/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetLightNodeSaleContracts(
	goCtx context.Context,
	_ *emptypb.Empty,
) (*types.QueryLightNodeSaleContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contracts, err := k.AllLightNodeSaleContracts(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryLightNodeSaleContractsResponse{
		LightNodeSaleContracts: contracts,
	}, nil
}
