package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidatorInfo returns validator info. It's not related to a snapshot.
func (k Keeper) ValidatorInfo(goCtx context.Context, req *types.QueryValidatorInfoRequest) (*types.QueryValidatorInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(req.ValAddr)
	if err != nil {
		return nil, err
	}

	externalAccounts, err := k.GetValidatorChainInfos(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryValidatorInfoResponse{
		ChainInfos: externalAccounts,
	}, nil
}
