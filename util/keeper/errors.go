package keeper

import (
	"errors"

	"github.com/VolumeFi/whoops"
)

const (
	ErrNotFound = whoops.Errorf("item (%T) not found in store: %s")
)

var ErrUnknownPanic = errors.New("unknown panic")
