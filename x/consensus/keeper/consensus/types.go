package consensus

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

const (
	consensusQueueIDCounterKey      = `consensus-queue-counter-`
	consensusBatchQueueIDCounterKey = `consensus-batch-queue-counter-`
	consensusQueueSigningKey        = `consensus-queue-signing-type-`

	consensusQueueMaxBatchSize = 100
)

type PutOptions struct {
	RequireSignatures bool
	PublicAccessData  []byte
}

//go:generate mockery --name=Queuer
type Queuer interface {
	Put(context.Context, ConsensusMsg, *PutOptions) (uint64, error)
	AddSignature(ctx context.Context, id uint64, signData *types.SignData) error
	AddEvidence(ctx context.Context, id uint64, evidence *types.Evidence) error
	SetPublicAccessData(ctx context.Context, id uint64, data *types.PublicAccessData) error
	GetPublicAccessData(ctx context.Context, id uint64) (*types.PublicAccessData, error)
	SetErrorData(ctx context.Context, id uint64, data *types.ErrorData) error
	GetErrorData(ctx context.Context, id uint64) (*types.ErrorData, error)
	Remove(context.Context, uint64) error
	GetAll(context.Context) ([]types.QueuedSignedMessageI, error)
	GetMsgByID(ctx context.Context, id uint64) (types.QueuedSignedMessageI, error)
	ChainInfo() (types.ChainType, string)
	ReassignValidator(ctx sdk.Context, id uint64, val string) error
}

type QueueBatcher interface {
	Queuer
	ProcessBatches(ctx context.Context) error
}

type SupportsConsensusQueueAction struct {
	QueueOptions
	ProcessMessageForAttestation func(ctx context.Context, q Queuer, msg types.QueuedSignedMessageI) error
}

type SupportsConsensusQueue interface {
	SupportedQueues(ctx context.Context) ([]SupportsConsensusQueueAction, error)
}
