package keeper

import (
	"github.com/palomachain/paloma/x/metrix/types"
)

var _ types.QueryServer = Keeper{}
