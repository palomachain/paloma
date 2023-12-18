package keeper

import (
	"github.com/VolumeFi/whoops"
)

const (
	ErrNotFound = whoops.Errorf("item (%T) not found in store: %s")
)
