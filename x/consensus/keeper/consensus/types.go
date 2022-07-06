package consensus

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

const (
	consensusQueueIDCounterKey      = `consensus-queue-counter-`
	consensusBatchQueueIDCounterKey = `consensus-batch-queue-counter-`
	consensusQueueSigningKey        = `consensus-queue-signing-type-`

	consensusQueueMaxBatchSize = 100
)

//go:generate mockery --name=Queuer
type Queuer interface {
	Put(sdk.Context, ...ConsensusMsg) error
	AddSignature(ctx sdk.Context, id uint64, signData *types.SignData) error
	AddEvidence(ctx sdk.Context, id uint64, evidence *types.Evidence) error
	SetPublicAccessData(ctx sdk.Context, id uint64, data *types.PublicAccessData) error
	GetPublicAccessData(ctx sdk.Context, id uint64) *types.PublicAccessData
	Remove(sdk.Context, uint64) error
	GetAll(sdk.Context) ([]types.QueuedSignedMessageI, error)
	GetMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error)
	ChainInfo() (types.ChainType, string)
}

type QueueBatcher interface {
	Queuer
	ProcessBatches(ctx sdk.Context) error
}

type SupportsConsensusQueueAction struct {
	QueueOptions

	Process func(ctx sdk.Context, q Queuer, msg types.QueuedSignedMessageI) error
}

type SupportsConsensusQueue interface {
	SupportedQueues(ctx sdk.Context) (map[string]SupportsConsensusQueueAction, error)
}
