package keeper

import (
	"context"

	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) DenomAuthorityMetadata(ctx context.Context, req *types.QueryDenomAuthorityMetadataRequest) (*types.QueryDenomAuthorityMetadataResponse, error) {
	authorityMetadata, err := k.GetAuthorityMetadata(ctx, req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomAuthorityMetadataResponse{AuthorityMetadata: authorityMetadata}, nil
}

func (k Keeper) DenomsFromCreator(ctx context.Context, req *types.QueryDenomsFromCreatorRequest) (*types.QueryDenomsFromCreatorResponse, error) {
	denoms := k.GetDenomsFromCreator(ctx, req.GetCreator())
	return &types.QueryDenomsFromCreatorResponse{Denoms: denoms}, nil
}
