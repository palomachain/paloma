package keeper

import (
	"github.com/palomachain/paloma/x/evm/types"
)

var _ types.QueryServer = Keeper{}
