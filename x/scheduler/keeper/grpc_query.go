package keeper

import (
	"github.com/palomachain/paloma/x/scheduler/types"
)

var _ types.QueryServer = Keeper{}
