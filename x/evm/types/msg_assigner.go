package types

import (
	"context"

	xchain "github.com/palomachain/paloma/internal/x-chain"
)

type MsgAssigner interface {
	PickValidatorForMessage(ctx context.Context, weights *RelayWeights, chainID string, requirements *xchain.JobRequirements) (string, error)
}
