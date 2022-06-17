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

	// TODO: check if ValidateBasic is being called somewhere in the Cosmos SDK
	// and if this is redundant call.
	// It should be caled on the "server" side as well as the client side of things.
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	err := k.AddSmartContractExecutionToConsensus(
		ctx,
		&types.Message{
			ChainID:     msg.ChainID,
			TurnstoneID: string(zero32Byte[:]),
			Action: &types.Message_SubmitLogicCall{
				SubmitLogicCall: &types.SubmitLogicCall{
					HexContractAddress: msg.GetHexSmartContractAddress(),
					Payload:            common.Hex2Bytes(msg.GetHexPayload()),
					Abi:                []byte(msg.GetAbi()),
					Deadline:           ctx.BlockTime().UTC().Add(5 * time.Minute).Unix(),
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitNewJobResponse{}, nil
}
