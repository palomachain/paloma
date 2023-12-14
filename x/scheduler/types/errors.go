package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/scheduler module sentinel errors
var (
	ErrSample               = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion       = sdkerrors.Register(ModuleName, 1501, "invalid version")

	ErrJobWithIDAlreadyExists = sdkerrors.Register(ModuleName, 1200, "job with id already exists")
	ErrJobNotFound            = sdkerrors.Register(ModuleName, 1201, "job not found")
	ErrInvalid                = sdkerrors.Register(ModuleName, 1202, "invalid")

	ErrCannotModifyJobPayload     = sdkerrors.Register(ModuleName, 1203, "cannot modify job's payload")
	ErrWasmExecuteMessageNotValid = sdkerrors.Register(ModuleName, 1204, "wasm message is not valid")
)
