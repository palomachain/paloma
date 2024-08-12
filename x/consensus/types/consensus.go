package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	xchain "github.com/palomachain/paloma/internal/x-chain"
)

const (
	cFlagGasEstimationRequired = 1 << 0 // Bit 0
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
	GetGasEstimates() []*GasEstimate
	GetGasEstimate() uint64
	AddSignData(*SignData)
	AddGasEstimate(*GasEstimate)
	SetElectedGasEstimate(uint64)
	AddEvidence(Evidence)
	GetEvidence() []*Evidence
	SetPublicAccessData(*PublicAccessData)
	GetPublicAccessData() *PublicAccessData
	SetErrorData(*ErrorData)
	GetErrorData() *ErrorData
	SetHandledAtBlockHeight(math.Int)
	GetHandledAtBlockHeight() *math.Int
	GetBytesToSign(AnyUnpacker) ([]byte, error)
	GetRequireSignatures() bool
	GetRequireGasEstimation() bool
	GetMsg() *types.Any
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

func BuildFlagMask(requireGasEstimation bool) uint32 {
	var value uint32 = 0
	if requireGasEstimation {
		value |= cFlagGasEstimationRequired
	}

	return value
}

var _ QueuedSignedMessageI = &QueuedSignedMessage{}

type MessageQueuedForBatchingI interface {
	proto.Message
	GetMsg() *types.Any
}

var _ MessageQueuedForBatchingI = &BatchOfConsensusMessages{}

func (q *QueuedSignedMessage) GetBytesToSign(unpacker AnyUnpacker) ([]byte, error) {
	msg, err := q.ConsensusMsg(unpacker)
	if err != nil {
		return nil, err
	}

	k := msg.(interface {
		Keccak256WithGasEstimate(uint64, uint64) []byte
	})

	return k.Keccak256WithGasEstimate(q.GetId(), q.GasEstimate), nil
}

func (q *QueuedSignedMessage) String() string {
	if q == nil {
		return ""
	}

	return fmt.Sprintf("%+v", *q)
}

func (q *QueuedSignedMessage) AddSignData(data *SignData) {
	if q.SignData == nil {
		q.SignData = []*SignData{}
	}
	q.SignData = append(q.SignData, data)
}

func (q *QueuedSignedMessage) AddGasEstimate(data *GasEstimate) {
	if q.GasEstimates == nil {
		q.GasEstimates = []*GasEstimate{}
	}
	q.GasEstimates = append(q.GasEstimates, data)
}

func (q *QueuedSignedMessage) SetElectedGasEstimate(estimate uint64) {
	// Setting an estimated gas value for a message
	// means we'll have to restart the signing process,
	// this time with the complete information.
	q.SignData = nil
	q.GasEstimate = estimate
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

func (q *QueuedSignedMessage) GetRequireGasEstimation() bool {
	return q.FlagMask&cFlagGasEstimationRequired != 0
}

func (q *QueuedSignedMessage) SetErrorData(data *ErrorData) {
	q.ErrorData = data
}

func (q *QueuedSignedMessage) GetHandledAtBlockHeight() *math.Int {
	return q.HandledAtBlockHeight
}

func (q *QueuedSignedMessage) SetHandledAtBlockHeight(i math.Int) {
	q.HandledAtBlockHeight = &i
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

// TODO should compute hash from msgs hashes
func (b *Batch) Keccak256WithGasEstimate(_, _ uint64) []byte {
	return nil
}

func Queue(queueTypeName string, typ xchain.Type, refID xchain.ReferenceID) string {
	return fmt.Sprintf("%s/%s/%s", typ, refID, queueTypeName)
}
