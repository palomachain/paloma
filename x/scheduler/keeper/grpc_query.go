package keeper

import (
	"github.com/palomachain/paloma/v2/x/scheduler/types"
)

var _ types.QueryServer = Keeper{}
