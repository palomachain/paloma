package keeper

import (
	"github.com/palomachain/paloma/v2/x/evm/types"
)

var _ types.QueryServer = Keeper{}
