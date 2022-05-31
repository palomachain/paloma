package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate mockery --name=AttestTask
type AttestTask interface {
	Attest()
}

type Evidence struct {
	From sdk.Address
	Data []byte
}

type AttestResult struct {
	// TODO
}

//go:generate mockery --name=Attestator
type Attestator interface {
	// ConsensusQueue tells in which ConsensusQueue to store the messages that require signatures.
	ConsensusQueue() ConsensusQueueType
	// Type tells the type of allowed message types for the ConsensusQueue.
	Type() any

	// Returns function which will be called with the underlying message and nonce
	// and it needs to return bytes for signing.
	BytesToSign() BytesToSignFunc

	// ValidateEvidence takes a task and an evidence and does a validation to make sure that it's correct.
	ValidateEvidence(ctx context.Context, task AttestTask, evidence Evidence) error
	// ProcessAllEvidence processes all given evidences and internally does whatever it needs to do with
	// that information. It returns the result back to the caller.
	ProcessAllEvidence(ctx context.Context, task AttestTask, evidence []Evidence) (AttestResult, error)
}
