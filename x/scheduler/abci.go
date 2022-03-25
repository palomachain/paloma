package scheduler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/scheduler/keeper"
)

func BeginBlocker(_ sdk.Context) {
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
}
