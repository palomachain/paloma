package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrValidatorWithAddrNotFound   = whoops.Errorf("validator with addr %s was not found")
	ErrMaxNumberOfExternalAccounts = whoops.Errorf("trying to submit %d accounts while the limit is %d")

	ErrSigningKeyNotFound             = whoops.Errorf("signing key for valAddr %s, chainType %s and chainID %s not found")
	ErrExternalChainAlreadyRegistered = whoops.Errorf("external account already registered: %s, %s")
)
