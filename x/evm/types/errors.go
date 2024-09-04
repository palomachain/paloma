package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/VolumeFi/whoops"
)

// x/evm module sentinel errors
var (
	ErrSample  = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalid = sdkerrors.Register(ModuleName, 1200, "invalid")
)

var (
	ErrEthTxNotVerified = whoops.String("transaction not verified")
	ErrEthTxFailed      = whoops.String("transaction failed to execute")
	ErrInvalidBalance   = whoops.Errorf("invalid balance: %s")
)
