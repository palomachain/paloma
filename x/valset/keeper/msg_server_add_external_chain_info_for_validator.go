package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types"
)

func (k msgServer) AddExternalChainInfoForValidator(goCtx context.Context, msg *types.MsgAddExternalChainInfoForValidator) (*types.MsgAddExternalChainInfoForValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	fmt.Println("AddExternalChainInfof>>>>>>>>>>>>>>")
	accAddr, err := sdk.AccAddressFromBech32(msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}
	fmt.Println("The accAddr is>>>>>>>>>>>>>.", accAddr)
	valAddr := sdk.ValAddress(accAddr)
	fmt.Printf("valAddr: >>>>>>>>>>>>>>>>>>%v\n", valAddr)
	fmt.Println("The chainsInfos are>>>>>>>>>>>>>>>", msg.ChainInfos)
	err = k.AddExternalChainInfo(ctx, valAddr, msg.ChainInfos)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddExternalChainInfoForValidatorResponse{}, nil
}
