package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/treasury module sentinel errors
var (
	ErrCannotAddZeroFunds     = sdkerrors.Register(ModuleName, 1100, "cannot add zero funds")
	ErrCannotAddNegativeFunds = sdkerrors.Register(ModuleName, 1101, "cannot add negative funds")
)
