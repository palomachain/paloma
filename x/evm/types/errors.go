package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/VolumeFi/whoops"
)

// x/evm module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1100, "sample error")
)

var (
	ErrEthTxNotVerified = whoops.String("transaction not verified")
	ErrInvalidBalance   = whoops.Errorf("invalid balance: %s")
)
