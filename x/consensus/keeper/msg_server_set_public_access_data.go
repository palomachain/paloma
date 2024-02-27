package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

func (k msgServer) SetPublicAccessData(goCtx context.Context, msg *types.MsgSetPublicAccessData) (*types.MsgSetPublicAccessDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, _ := sdk.AccAddressFromBech32(msg.Metadata.Creator)

	valAddr := sdk.ValAddress(creator.Bytes())

	if err := k.Keeper.valset.CanAcceptValidator(ctx, valAddr); err != nil {
		return nil, err
	}

	err := k.Keeper.SetMessagePublicAccessData(ctx, valAddr, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetPublicAccessDataResponse{}, nil
}
