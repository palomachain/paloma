package keeper

import (
	"github.com/palomachain/paloma/v2/x/paloma/types"
)

var _ types.QueryServer = Keeper{}
