package keeper

import (
	"github.com/palomachain/paloma/v2/x/valset/types"
)

var _ types.QueryServer = Keeper{}
