package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrValidatorWithAddrNotFound   = whoops.Errorf("validator with addr %s was not found")
	ErrMaxNumberOfExternalAccounts = whoops.Errorf("trying to submit %d accounts while the limit is %d")
)
