package keeper

import (
	"context"

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

	// TODO: use real chain IDs
	err := k.AddSmartContractExecutionToConsensus(
		ctx,
		msg.GetChainType(),
		msg.GetChainID(),
		&types.ArbitrarySmartContractCall{
			HexAddress: msg.HexSmartContractAddress,
			Payload:    common.Hex2Bytes(msg.GetHexPayload()),
			Abi:        []byte(msg.GetAbi()),
		})

	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitNewJobResponse{}, nil
}
