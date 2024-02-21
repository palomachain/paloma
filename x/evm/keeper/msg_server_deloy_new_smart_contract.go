package keeper

import (
	"context"

	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) DeployNewSmartContract(goCtx context.Context, msg *types.MsgDeployNewSmartContractRequest) (*types.DeployNewSmartContractResponse, error) {
	ctx := common.SdkContext(goCtx)
	logger := ctx.Logger()

	logger.Info("TODO: Implement me")

	return &types.DeployNewSmartContractResponse{}, nil
}
