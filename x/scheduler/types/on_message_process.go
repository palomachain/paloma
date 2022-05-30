package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type OnMessageProcessor interface {
	OnSchedulerMessageProcess(ctx sdk.Context, rawMsg any) (processed bool, err error)
}
