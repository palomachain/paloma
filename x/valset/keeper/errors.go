package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrValidatorAlreadyRegistered = whoops.String("validator is already registered")
	ErrValidatorWithAddrNotFound  = whoops.Errorf("validator with addr %s was not found")
)
