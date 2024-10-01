package keeper

import (
	"github.com/palomachain/paloma/v2/x/consensus/types"
)

var _ types.QueryServer = Keeper{}
