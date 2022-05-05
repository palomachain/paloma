package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

//go:generate mockery --name=QueuedSignedMessageI
type QueuedSignedMessageI interface {
	proto.Message
	GetId() uint64
	Nonce() []byte
	GetMsg() *codectypes.Any
	SdkMsg() (sdk.Msg, error)
	GetSignData() []*SignData
	AddSignData(*SignData)
}

var _ QueuedSignedMessageI = &QueuedSignedMessage{}

func (q *QueuedSignedMessage) AddSignData(data *SignData) {
	if q.SignData == nil {
		q.SignData = []*SignData{}
	}
	q.SignData = append(q.SignData, data)
}

func (q *QueuedSignedMessage) Nonce() []byte {
	return sdk.Uint64ToBigEndian(q.GetId())
}

func (q *QueuedSignedMessage) SdkMsg() (sdk.Msg, error) {
	var sdkMsg sdk.Msg
	if err := ModuleCdc.UnpackAny(q.Msg, &sdkMsg); err != nil {
		return nil, err
	}
	return sdkMsg, nil
}
