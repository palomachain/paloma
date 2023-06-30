package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type MsgAssigner interface {
	PickValidatorForMessage(ctx sdk.Context, weights RelayWeights) (string, error)
}
