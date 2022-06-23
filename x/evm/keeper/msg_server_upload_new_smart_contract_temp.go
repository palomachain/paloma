package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k msgServer) UploadNewSmartContractTemp(goCtx context.Context, msg *types.MsgUploadNewSmartContractTemp) (*types.MsgUploadNewSmartContractTempResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.addUploadSmartContractToConsensus(ctx, msg.GetChainID(), &types.UploadSmartContract{
		Bytecode:         common.FromHex(msg.GetBytecode()),
		Abi:              []byte(msg.GetAbi()),
		ConstructorInput: common.FromHex(msg.GetConstructorInput()),
	})

	if err != nil {
		return nil, err
	}

	return &types.MsgUploadNewSmartContractTempResponse{}, nil
}
