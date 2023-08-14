package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
)

type MsgAssigner interface {
	PickValidatorForMessage(ctx sdk.Context, weights *RelayWeights, chainID string, requirements *xchain.JobRequirements) (string, error)
}
