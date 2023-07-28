package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/bech32ibc module sentinel errors
// ibc-go registers error codes under 10
var (
	ErrInvalidHRP                 = sdkerrors.Register(ModuleName, 11, "Invalid HRP")
	ErrInvalidIBCData             = sdkerrors.Register(ModuleName, 12, "Invalid IBC Data")
	ErrRecordNotFound             = sdkerrors.Register(ModuleName, 13, "No record found for requested HRP")
	ErrNoNativeHrp                = sdkerrors.Register(ModuleName, 14, "No native prefix was set")
	ErrInvalidOffsetHeightTimeout = sdkerrors.Register(ModuleName, 15, "At least one of offset height or offset timeout should be set")
)
