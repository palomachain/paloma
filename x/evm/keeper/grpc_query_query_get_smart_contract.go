package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryGetSmartContract(goCtx context.Context, req *types.QueryQueryGetSmartContractRequest) (*types.QueryQueryGetSmartContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var smartContract *types.SmartContract
	var err error

	if req.GetSmartContractID() == 0 {
		smartContract, err = k.GetLastSmartContract(ctx)
	} else {
		smartContract, err = k.getSmartContract(ctx, req.GetSmartContractID())
	}

	if err != nil {
		return nil, err
	}

	return &types.QueryQueryGetSmartContractResponse{
		ID:       smartContract.GetId(),
		Abi:      smartContract.GetAbiJSON(),
		Bytecode: smartContract.GetBytecode(),
	}, nil
}
