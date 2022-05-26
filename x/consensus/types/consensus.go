package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

type ConsensusQueueType string

//go:generate mockery --name=QueuedSignedMessageI
type QueuedSignedMessageI interface {
	proto.Message
	GetId() uint64
	Nonce() []byte
	ConsensusMsg() (ConsensusMsg, error)
	GetSignData() []*SignData
	AddSignData(*SignData)
	GetBytesToSign() []byte
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

func (q *QueuedSignedMessage) ConsensusMsg() (ConsensusMsg, error) {
	var ptr ConsensusMsg
	if err := ModuleCdc.UnpackAny(q.Msg, &ptr); err != nil {
		return nil, err
	}
	return ptr, nil
}

func (b *Batch) GetSignBytes() []byte {
	return b.GetBytesToSign()
}
