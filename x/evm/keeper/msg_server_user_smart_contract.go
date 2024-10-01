package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/evm/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (k msgServer) UploadUserSmartContract(
	ctx context.Context,
	req *types.MsgUploadUserSmartContractRequest,
) (*types.MsgUploadUserSmartContractResponse, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(req.Metadata.Creator)
	if err != nil {
		return nil, err
	}

	valAddress := sdk.ValAddress(creatorAddr.Bytes()).String()

	contract := &types.UserSmartContract{
		Author:           valAddress,
		Title:            req.Title,
		AbiJson:          req.AbiJson,
		Bytecode:         req.Bytecode,
		ConstructorInput: req.ConstructorInput,
	}

	id, err := k.SaveUserSmartContract(ctx, valAddress, contract)
	if err != nil {
		return nil, err
	}

	return &types.MsgUploadUserSmartContractResponse{
		Id: id,
	}, nil
}

func (k msgServer) RemoveUserSmartContract(
	ctx context.Context,
	req *types.MsgRemoveUserSmartContractRequest,
) (*emptypb.Empty, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(req.Metadata.Creator)
	if err != nil {
		return nil, err
	}

	valAddress := sdk.ValAddress(creatorAddr.Bytes()).String()

	err = k.DeleteUserSmartContract(ctx, valAddress, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (k msgServer) DeployUserSmartContract(
	ctx context.Context,
	req *types.MsgDeployUserSmartContractRequest,
) (*types.MsgDeployUserSmartContractResponse, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(req.Metadata.Creator)
	if err != nil {
		return nil, err
	}

	valAddress := sdk.ValAddress(creatorAddr.Bytes()).String()

	id, err := k.CreateUserSmartContractDeployment(ctx, valAddress, req.Id,
		req.TargetChain)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeployUserSmartContractResponse{
		MsgId: id,
	}, nil
}
