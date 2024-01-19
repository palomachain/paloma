package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var ErrInvalidRequest = sdkerrors.Register(ModuleName, 1, "invalid request")
