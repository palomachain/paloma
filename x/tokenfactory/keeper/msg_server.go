package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreateDenom(ctx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	denom, err := server.Keeper.CreateDenom(ctx, msg.Metadata.Creator, msg.Subdenom)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgCreateDenom,
			sdk.NewAttribute(types.AttributeCreator, msg.Metadata.Creator),
			sdk.NewAttribute(types.AttributeNewTokenDenom, denom),
		),
	})

	return &types.MsgCreateDenomResponse{
		NewTokenDenom: denom,
	}, nil
}

func (server msgServer) Mint(ctx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	_, denomExists := server.bankKeeper.GetDenomMetaData(ctx, msg.Amount.Denom)
	if !denomExists {
		return nil, types.ErrDenomDoesNotExist.Wrapf("denom: %s", msg.Amount.Denom)
	}

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Metadata.Creator != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	err = server.Keeper.mintTo(ctx, msg.Amount, msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgMint,
			sdk.NewAttribute(types.AttributeMintToAddress, msg.Metadata.Creator),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
		),
	})

	return &types.MsgMintResponse{}, nil
}

func (server msgServer) Burn(ctx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Metadata.Creator != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	err = server.Keeper.burnFrom(ctx, msg.Amount, msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgBurn,
			sdk.NewAttribute(types.AttributeBurnFromAddress, msg.Metadata.Creator),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
		),
	})

	return &types.MsgBurnResponse{}, nil
}

func (server msgServer) ChangeAdmin(ctx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	if msg.Metadata.Creator != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	err = server.Keeper.setAdmin(ctx, msg.Denom, msg.NewAdmin)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgChangeAdmin,
			sdk.NewAttribute(types.AttributeDenom, msg.GetDenom()),
			sdk.NewAttribute(types.AttributeNewAdmin, msg.NewAdmin),
		),
	})

	return &types.MsgChangeAdminResponse{}, nil
}

func (server msgServer) SetDenomMetadata(ctx context.Context, msg *types.MsgSetDenomMetadata) (*types.MsgSetDenomMetadataResponse, error) {
	err := msg.DenomMetadata.Validate()
	if err != nil {
		return nil, err
	}

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.DenomMetadata.Base)
	if err != nil {
		return nil, err
	}

	if msg.Metadata.Creator != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	server.Keeper.bankKeeper.SetDenomMetaData(ctx, msg.DenomMetadata)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetDenomMetadata,
			sdk.NewAttribute(types.AttributeDenom, msg.DenomMetadata.Base),
			sdk.NewAttribute(types.AttributeDenomMetadata, msg.Metadata.String()),
		),
	})

	return &types.MsgSetDenomMetadataResponse{}, nil
}

func (server msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, types.ErrInvalidParams
	}

	if msg.Authority != msg.Metadata.Creator {
		return nil, types.ErrUnauthorized.Wrapf("authority mismatch: %s", msg.Authority)
	}

	if msg.Authority != server.authority {
		return nil, types.ErrUnauthorized.Wrapf("unexpected authority: %s", msg.Authority)
	}

	server.SetParams(ctx, msg.Params)
	return &types.MsgUpdateParamsResponse{}, nil
}
