package keeper

import (
	"github.com/palomachain/paloma/x/valset/types"
)

var _ types.QueryServer = Keeper{}
