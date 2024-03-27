package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MessageByID(goCtx context.Context, req *types.QueryMessageByIDRequest) (*types.MessageWithSignatures, error) {
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
	approvedMessage, err := toMessageWithSignatures(msg, k.cdc)
	if err != nil {
		return nil, err
	}
	approvedMessage.Evidence = append(approvedMessage.Evidence, msg.GetEvidence()...)
	return &approvedMessage, nil
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

		approvedMessage, err := toMessageWithSignatures(msg, k.cdc)
		if err != nil {
			return nil, err
		}
		res.Messages = append(res.Messages, approvedMessage)
	}
	return res, nil
}

func toMessageWithSignatures(msg types.QueuedSignedMessageI, cdc codec.BinaryCodec) (types.MessageWithSignatures, error) {
	origMsg, err := msg.ConsensusMsg(cdc)
	if err != nil {
		return types.MessageWithSignatures{}, err
	}
	anyMsg, err := codectypes.NewAnyWithValue(origMsg)
	if err != nil {
		return types.MessageWithSignatures{}, err
	}

	var publicAccessData []byte

	if msg.GetPublicAccessData() != nil {
		publicAccessData = msg.GetPublicAccessData().GetData()
	}

	var errorData []byte

	if msg.GetErrorData() != nil {
		errorData = msg.GetErrorData().GetData()
	}

	approvedMessage := types.MessageWithSignatures{
		Nonce:            msg.Nonce(),
		Id:               msg.GetId(),
		Msg:              anyMsg,
		BytesToSign:      msg.GetBytesToSign(),
		SignData:         []*types.ValidatorSignature{},
		PublicAccessData: publicAccessData,
		ErrorData:        errorData,
	}
	for _, signData := range msg.GetSignData() {
		approvedMessage.SignData = append(approvedMessage.SignData, &types.ValidatorSignature{
			ValAddress:             signData.GetValAddress(),
			Signature:              signData.GetSignature(),
			ExternalAccountAddress: signData.GetExternalAccountAddress(),
			PublicKey:              signData.GetPublicKey(),
		})
	}

	return approvedMessage, nil
}
