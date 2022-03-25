package keeper

import (
	"github.com/volumefi/cronchain/x/scheduler/types"
)

var _ types.QueryServer = Keeper{}
