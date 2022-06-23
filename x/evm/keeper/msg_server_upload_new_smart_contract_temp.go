package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k msgServer) UploadNewSmartContractTemp(goCtx context.Context, msg *types.MsgUploadNewSmartContractTemp) (*types.MsgUploadNewSmartContractTempResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.addUploadSmartContractToConsensus(ctx, msg.GetChainID(), &types.UploadSmartContract{
		Bytecode:         []byte(msg.GetBytecode()),
		Abi:              []byte(msg.GetAbi()),
		ConstructorInput: []byte(msg.GetConstructorInput()),
	})

	if err != nil {
		return nil, err
	}

	return &types.MsgUploadNewSmartContractTempResponse{}, nil
}
