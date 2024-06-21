package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/evm/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k Keeper) UploadUserSmartContract(
	ctx context.Context,
	req *types.MsgUploadUserSmartContractRequest,
) (*types.MsgUploadUserSmartContractResponse, error) {
	contract := &types.UserSmartContract{
		ValAddress:       req.ValAddress,
		Title:            req.Title,
		AbiJson:          req.AbiJson,
		Bytecode:         req.Bytecode,
		ConstructorInput: req.ConstructorInput,
	}

	id, err := k.SaveUserSmartContract(ctx, req.ValAddress, contract)
	if err != nil {
		return nil, err
	}

	return &types.MsgUploadUserSmartContractResponse{
		Id: id,
	}, nil
}

func (k Keeper) RemoveUserSmartContract(
	ctx context.Context,
	req *types.MsgRemoveUserSmartContractRequest,
) (*emptypb.Empty, error) {
	err := k.DeleteUserSmartContract(ctx, req.ValAddress, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
