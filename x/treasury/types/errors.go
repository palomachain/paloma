package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/treasury module sentinel errors
var (
	ErrCannotAddZeroFunds     = sdkerrors.Register(ModuleName, 1100, "cannot add zero funds")
	ErrCannotAddNegativeFunds = sdkerrors.Register(ModuleName, 1101, "cannot add negative funds")
)
