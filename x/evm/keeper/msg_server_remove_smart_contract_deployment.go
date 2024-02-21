package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) RemoveSmartContractDeployment(ctx context.Context, req *types.MsgRemoveSmartContractDeploymentRequest) (*types.RemoveSmartContractDeploymentResponse, error) {
	k.DeleteSmartContractDeploymentByContractID(common.SdkContext(ctx), req.SmartContractID, req.ChainReferenceID)
	return &types.RemoveSmartContractDeploymentResponse{}, nil
}
