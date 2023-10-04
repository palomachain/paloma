package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) RemoveSmartContractDeployment(ctx context.Context, req *types.MsgRemoveSmartContractDeploymentRequest) (*types.RemoveSmartContractDeploymentResponse, error) {
	k.DeleteSmartContractDeploymentByContractID(sdk.UnwrapSDKContext(ctx), req.SmartContractID, req.ChainReferenceID)
	return &types.RemoveSmartContractDeploymentResponse{}, nil
}
