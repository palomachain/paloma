package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/evm/types"
)

func (k msgServer) UploadNewSmartContractTemp(goCtx context.Context, msg *types.MsgUploadNewSmartContractTemp) (*types.MsgUploadNewSmartContractTempResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// err := k.DeployNewSmartContract(ctx, msg.GetChainReferenceID(), &types.UploadSmartContract{
	// 	Bytecode:         common.FromHex(msg.GetBytecode()),
	// 	Abi:              []byte(msg.GetAbi()),
	// 	ConstructorInput: common.FromHex(msg.GetConstructorInput()),
	// })

	// if err != nil {
	// 	return nil, err
	// }

	return &types.MsgUploadNewSmartContractTempResponse{}, nil
}
