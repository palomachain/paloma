package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MessageByID(goCtx context.Context, req *types.QueryMessageByIDRequest) (*types.MessageQueryResult, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cq, err := k.getConsensusQueue(ctx, req.QueueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue", "err", err)
		return nil, err
	}
	msg, err := cq.GetMsgByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, fmt.Errorf("message not found")
	}
	approvedMessage, err := consensus.ToMessageWithSignatures(msg, k.cdc)
	if err != nil {
		return nil, err
	}
	res := &types.MessageQueryResult{
		Message: &approvedMessage,
	}

	copy(res.Evidence, msg.GetEvidence())
	copy(res.GasEstimates, msg.GetGasEstimates())

	return res, nil
}

func (k Keeper) MessagesInQueue(goCtx context.Context, req *types.QueryMessagesInQueueRequest) (*types.QueryMessagesInQueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesFromQueue(ctx, req.QueueTypeName, 0)
	if err != nil {
		return nil, err
	}

	res := &types.QueryMessagesInQueueResponse{}
	skipIfValidatorProvidedEvidence := req.GetSkipEvidenceProvidedByValAddress()
	for _, msg := range msgs {
		if skipIfValidatorProvidedEvidence != nil {
			shouldSkipThisMsg := false
			for _, evidence := range msg.GetEvidence() {
				if evidence.ValAddress.Equals(skipIfValidatorProvidedEvidence) {
					shouldSkipThisMsg = true
					break
				}
			}

			if shouldSkipThisMsg {
				continue
			}
		}

		approvedMessage, err := consensus.ToMessageWithSignatures(msg, k.cdc)
		if err != nil {
			return nil, err
		}
		res.Messages = append(res.Messages, approvedMessage)
	}
	return res, nil
}
