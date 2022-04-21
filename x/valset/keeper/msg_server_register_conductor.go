package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/valset/types"
)

func (k msgServer) RegisterConductor(goCtx context.Context, msg *types.MsgRegisterConductor) (*types.MsgRegisterConductorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Register(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgRegisterConductorResponse{}, nil
}
