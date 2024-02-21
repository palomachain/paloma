package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) DeployNewSmartContract(goCtx context.Context, msg *types.MsgDeployNewSmartContractRequest) (*types.DeployNewSmartContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger()

	logger.Info("TODO: Implement me")

	return &types.DeployNewSmartContractResponse{}, nil
}
