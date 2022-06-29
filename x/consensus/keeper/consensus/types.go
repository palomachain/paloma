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
	Remove(sdk.Context, uint64) error
	GetAll(sdk.Context) ([]types.QueuedSignedMessageI, error)
	GetMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error)
	ChainInfo() (chainType types.ChainType, chainID string)
	ConsensusQueue() string
}

type QueueBatcher interface {
	Queuer
	ProcessBatches(ctx sdk.Context) error
}

type SupportsConsensusQueue interface {
	SupportsQueue(ctx sdk.Context, queueTypeName string) (*QueueOptions, error)
}
