package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/consensus/types"
)

func (k msgServer) SetErrorData(goCtx context.Context, msg *types.MsgSetErrorData) (*types.MsgSetErrorDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, _ := sdk.AccAddressFromBech32(msg.Metadata.Creator)

	valAddr := sdk.ValAddress(creator.Bytes())

	if err := k.Keeper.valset.CanAcceptValidator(ctx, valAddr); err != nil {
		return nil, err
	}

	if len(msg.Data) < 1 {
		return nil, fmt.Errorf("data must not be nil")
	}

	err := k.Keeper.SetMessageErrorData(ctx, valAddr, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetErrorDataResponse{}, nil
}
