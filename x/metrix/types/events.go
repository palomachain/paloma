package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OnConsensusMessageAttestedListener interface {
	OnConsensusMessageAttested(context.Context, MessageAttestedEvent)
}

type MessageAttestedEvent struct {
	AssignedAtBlockHeight  math.Int
	HandledAtBlockHeight   math.Int
	Assignee               sdk.ValAddress
	MessageID              uint64
	WasRelayedSuccessfully bool
}
