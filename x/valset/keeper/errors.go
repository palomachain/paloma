package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrValidatorAlreadyRegistered     = whoops.String("validator is already registered")
	ErrValidatorWithAddrNotFound      = whoops.Errorf("validator with addr %s was not found")
	ErrPublicKeyOrSignatureIsInvalid  = whoops.String("public key or signature is invalid. couldn't validate the signature.")
	ErrExternalChainAlreadyRegistered = whoops.Errorf("external chain info for chain id %s with address %s already exists")
)
