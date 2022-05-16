package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

type AttestTask interface {
	GetTaskDescription() []byte
	Attest()
}

type Evidence struct {
	From sdk.Address
	Data []byte
}

type AttestResult struct {
	// TODO
}

type Attestator interface {
	ConsensusQueue() types.ConsensusQueueType
	ValidateEvidence(ctx context.Context, task AttestTask, evidence Evidence) error
	ProcessAllEvidence(ctx context.Context, task AttestTask, evidence []Evidence) (AttestResult, error)
}
