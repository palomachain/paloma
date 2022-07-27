package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k msgServer) SubmitNewJob(goCtx context.Context, msg *types.MsgSubmitNewJob) (*types.MsgSubmitNewJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: remove this
	panic("do not use this")

	err := k.AddSmartContractExecutionToConsensus(
		ctx,
		msg.GetChainReferenceID(),
		string(zero32Byte[:]),
		&types.SubmitLogicCall{
			HexContractAddress: msg.GetHexSmartContractAddress(),
			Payload:            common.Hex2Bytes(msg.GetHexPayload()),
			Abi:                []byte(msg.GetAbi()),
			Deadline:           ctx.BlockTime().UTC().Add(5 * time.Minute).Unix(),
		},
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitNewJobResponse{}, nil
}
