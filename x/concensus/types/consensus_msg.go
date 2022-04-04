package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

type QueuedSignedMessageI interface {
	proto.Message
	GetId() uint64
	SdkMsg() (sdk.Msg, error)
	GetSigners() []*Signer
	AddSignature(*Signer)
}

var _ QueuedSignedMessageI = &QueuedSignedMessage{}

func (q *QueuedSignedMessage) AddSignature(sig *Signer) {
	if q.Signers == nil {
		q.Signers = []*Signer{}
	}
	q.Signers = append(q.Signers, sig)
}

func (q *QueuedSignedMessage) SdkMsg() (sdk.Msg, error) {
	var sdkMsg sdk.Msg
	if err := ModuleCdc.UnpackAny(q.Msg, &sdkMsg); err != nil {
		return nil, err
	}
	return sdkMsg, nil
}
