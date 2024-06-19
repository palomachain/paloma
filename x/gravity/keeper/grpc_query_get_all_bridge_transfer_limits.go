package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) GetAllBridgeTransferLimits(
	goCtx context.Context,
	_ *emptypb.Empty,
) (*types.QueryAllBridgeTransferLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	limits, err := k.AllBridgeTransferLimits(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryAllBridgeTransferLimitsResponse{
		BridgeTransferLimits: limits,
	}, nil
}
