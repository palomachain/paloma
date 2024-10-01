package scheduler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/scheduler/keeper"
)

func BeginBlocker(_ sdk.Context) {
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
}
