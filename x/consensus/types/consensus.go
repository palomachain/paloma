package types

import (
	"fmt"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"time"
)

type ConsensusQueueType string

//go:generate mockery --name=QueuedSignedMessageI
type QueuedSignedMessageI interface {
	proto.Message
	GetId() uint64
	Nonce() []byte
	GetAddedAtBlockHeight() int64
	GetAddedAt() time.Time
	ConsensusMsg(AnyUnpacker) (ConsensusMsg, error)
	GetSignData() []*SignData
	AddSignData(*SignData)
	AddEvidence(Evidence)
	GetEvidence() []*Evidence
	SetPublicAccessData(*PublicAccessData)
	GetPublicAccessData() *PublicAccessData
	GetBytesToSign() []byte
	GetRequireSignatures() bool
}

type Salt struct {
	Nonce     uint64
	ExtraData []byte
}

type BytesToSignFunc func(msg ConsensusMsg, salt Salt) []byte

type VerifySignatureFunc func(msg []byte, sig []byte, pk []byte) bool

func TypedBytesToSign[T any](fnc func(msg T, salt Salt) []byte) BytesToSignFunc {
	return BytesToSignFunc(func(raw ConsensusMsg, salt Salt) []byte {
		msgT, ok := raw.(T)
		if !ok {
			var expected T
			panic(fmt.Sprintf("can't process message of type: %T, expected: %T", raw, expected))
		}
		return fnc(msgT, salt)
	})
}

var _ QueuedSignedMessageI = &QueuedSignedMessage{}

type MessageQueuedForBatchingI interface {
	proto.Message
	GetMsg() *types.Any
}

var _ MessageQueuedForBatchingI = &BatchOfConsensusMessages{}

func (q *QueuedSignedMessage) AddSignData(data *SignData) {
	if q.SignData == nil {
		q.SignData = []*SignData{}
	}
	q.SignData = append(q.SignData, data)
}

func (q *QueuedSignedMessage) AddEvidence(data Evidence) {
	if q.Evidence == nil {
		q.Evidence = []*Evidence{}
	}

	for i := range q.Evidence {
		if q.Evidence[i].ValAddress.Equals(data.ValAddress) {
			q.Evidence[i].Proof = data.Proof
			return
		}
	}

	// if validator didn't provide evidence already, then add it here for the first time
	q.Evidence = append(q.Evidence, &data)
}

func (q *QueuedSignedMessage) SetPublicAccessData(data *PublicAccessData) {
	q.PublicAccessData = data
}

func (q *QueuedSignedMessage) Nonce() []byte {
	return sdk.Uint64ToBigEndian(q.GetId())
}

func (q *QueuedSignedMessage) ConsensusMsg(unpacker AnyUnpacker) (ConsensusMsg, error) {
	var ptr ConsensusMsg
	if err := unpacker.UnpackAny(q.Msg, &ptr); err != nil {
		return nil, err
	}
	return ptr, nil
}

func (b *Batch) GetSignBytes() []byte {
	return b.GetBytesToSign()
}

func Queue(queueTypeName string, chainType ChainType, chainReferenceID string) string {
	return fmt.Sprintf("%s/%s/%s", chainType, chainReferenceID, queueTypeName)
}
