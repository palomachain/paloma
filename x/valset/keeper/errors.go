package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrValidatorWithAddrNotFound   = whoops.Errorf("validator with addr %s was not found")
	ErrMaxNumberOfExternalAccounts = whoops.Errorf("trying to submit %d accounts while the limit is %d")
	ErrValidatorCannotBePigeon     = whoops.Errorf("validator %s cannot be a pigeon")

	ErrSigningKeyNotFound             = whoops.Errorf("signing key for valAddr %s, chainType %s and chainReferenceID %s not found")
	ErrExternalChainAlreadyRegistered = whoops.Errorf("external account already registered: %s, %s, %s. Existing owner: %s, New owner: %s")
)
