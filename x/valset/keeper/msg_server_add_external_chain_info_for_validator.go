package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/valset/types"
)

func (k msgServer) AddExternalChainInfoForValidator(goCtx context.Context, msg *types.MsgAddExternalChainInfoForValidator) (*types.MsgAddExternalChainInfoForValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.addExternalChainInfo(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddExternalChainInfoForValidatorResponse{}, nil
}
