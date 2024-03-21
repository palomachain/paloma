package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
)

func (k msgServer) AddExternalChainInfoForValidator(goCtx context.Context, msg *types.MsgAddExternalChainInfoForValidator) (*types.MsgAddExternalChainInfoForValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accAddr, err := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}
	valAddr := sdk.ValAddress(accAddr)
	err = k.AddExternalChainInfo(ctx, valAddr, msg.ChainInfos)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddExternalChainInfoForValidatorResponse{}, nil
}
