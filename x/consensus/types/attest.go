package types

import "context"

//go:generate mockery --name=AttestTask
type AttestTask interface {
	Attest()
}

// type Evidence struct {
// 	From sdk.Address
// 	Data []byte
// }

type AttestResult struct {
	// TODO
}

//go:generate mockery --name=Attestator
type Attestator interface {
	// ValidateEvidence takes a task and an evidence and does a validation to make sure that it's correct.
	ValidateEvidence(ctx context.Context, task AttestTask, evidence Evidence) error
	// ProcessAllEvidence processes all given evidences and internally does whatever it needs to do with
	// that information. It returns the result back to the caller.
	ProcessAllEvidence(ctx context.Context, task AttestTask, evidence []Evidence) (AttestResult, error)
}
