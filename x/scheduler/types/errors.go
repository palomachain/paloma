package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/scheduler module sentinel errors
var (
	ErrSample               = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion       = sdkerrors.Register(ModuleName, 1501, "invalid version")

	ErrJobWithIDAlreadyExists = sdkerrors.Register(ModuleName, 1, "job with id already exists")
	ErrJobNotFound            = sdkerrors.Register(ModuleName, 2, "job not found")
)
