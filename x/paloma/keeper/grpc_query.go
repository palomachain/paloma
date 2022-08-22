package keeper

import (
	"github.com/palomachain/paloma/x/paloma/types"
)

var _ types.QueryServer = Keeper{}
