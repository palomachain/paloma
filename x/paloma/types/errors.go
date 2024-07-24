package types

// DONTCOVER

import (
	"errors"

	sdkerrors "cosmossdk.io/errors"
)

// x/paloma module sentinel errors
var (
	ErrSample            = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrAccountExists     = errors.New("account already exists")
	ErrLicenseExists     = errors.New("license already exists")
	ErrNoAccount         = errors.New("account not found")
	ErrNoFeegranter      = errors.New("no feegranter set")
	ErrNoFunder          = errors.New("no funder set")
	ErrNoLicense         = errors.New("no license found for this client")
)
