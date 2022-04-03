package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueuedSignedMessageI interface {
	SdkMsg() (sdk.Msg, error)
	GetSigners() []*Signer
	AddSignature(*Signer) error
}

var _ QueuedSignedMessageI = &QueuedSignedMessage{}

func (q *QueuedSignedMessage) AddSignature(sig *Signer) error {
	return nil
}

func (q *QueuedSignedMessage) SdkMsg() (sdk.Msg, error) {
	var sdkMsg sdk.Msg
	if err := ModuleCdc.UnpackAny(q.Msg, &sdkMsg); err != nil {
		return nil, err
	}
	return sdkMsg, nil
}
