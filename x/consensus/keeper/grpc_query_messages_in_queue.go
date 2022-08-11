package keeper

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MessagesInQueue(goCtx context.Context, req *types.QueryMessagesInQueueRequest) (*types.QueryMessagesInQueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := k.GetMessagesFromQueue(ctx, req.QueueTypeName, 200)
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

		origMsg, err := msg.ConsensusMsg(k.cdc)

		if err != nil {
			return nil, err
		}
		anyMsg, err := codectypes.NewAnyWithValue(origMsg)
		if err != nil {
			return nil, err
		}

		var publicAccessData []byte

		if msg.GetPublicAccessData() != nil {
			publicAccessData = msg.GetPublicAccessData().GetData()
		}
		approvedMessage := &types.MessageWithSignatures{
			Nonce:            msg.Nonce(),
			Id:               msg.GetId(),
			Msg:              anyMsg,
			BytesToSign:      msg.GetBytesToSign(),
			SignData:         []*types.ValidatorSignature{},
			PublicAccessData: publicAccessData,
		}
		for _, signData := range msg.GetSignData() {
			approvedMessage.SignData = append(approvedMessage.SignData, &types.ValidatorSignature{
				ValAddress:             signData.GetValAddress(),
				Signature:              signData.GetSignature(),
				ExternalAccountAddress: signData.GetExternalAccountAddress(),
				PublicKey:              signData.GetPublicKey(),
			})
		}
		res.Messages = append(res.Messages, approvedMessage)
	}
	return res, nil
}
