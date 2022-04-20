package keeper

import (
	"github.com/volumefi/cronchain/x/valset/types"
)

var _ types.QueryServer = Keeper{}
