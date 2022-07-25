package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrNotFound = whoops.Errorf("item (%T) not found in store: %s")
)
