package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k msgServer) SubmitNewJob(goCtx context.Context, msg *types.MsgSubmitNewJob) (*types.MsgSubmitNewJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: use real chain IDs
	err := k.AddSmartContractExecutionToConsensus(ctx, &types.ArbitrarySmartContractCall{
		HexAddress: msg.HexSmartContractAddress,
		ChainID:    "test",
		Payload:    common.Hex2Bytes(msg.GetHexPayload()),
		Abi:        []byte(msg.GetAbi()),
	})

	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitNewJobResponse{}, nil
}
