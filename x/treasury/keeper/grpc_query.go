package keeper

import (
	"github.com/palomachain/paloma/x/treasury/types"
)

var _ types.QueryServer = Keeper{}
