package keeper

import (
	"github.com/volumefi/cronchain/x/consensus/types"
)

var _ types.QueryServer = Keeper{}
