package keeper

import (
	"github.com/volumefi/cronchain/x/concensus/types"
)

var _ types.QueryServer = Keeper{}
