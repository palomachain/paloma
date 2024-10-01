package types

import (
	"context"

	xchain "github.com/palomachain/paloma/v2/internal/x-chain"
)

type MsgAssigner interface {
	PickValidatorForMessage(ctx context.Context, weights *RelayWeights, chainID string, requirements *xchain.JobRequirements) (string, error)
}
