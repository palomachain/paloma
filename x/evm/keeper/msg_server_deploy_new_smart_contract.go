package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) DeployNewSmartContract(
	ctx context.Context,
	msg *types.MsgDeployNewSmartContractRequest,
) (*types.DeployNewSmartContractResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := sdkCtx.Logger()

	logger.Info("TODO: Implement me")

	return &types.DeployNewSmartContractResponse{}, nil
}
