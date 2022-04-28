package keeper

import (
	"github.com/palomachain/paloma/x/consensus/types"
)

var _ types.QueryServer = Keeper{}
