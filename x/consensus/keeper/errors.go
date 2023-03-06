package keeper

import (
	"github.com/VolumeFi/whoops"
)

const (
	ErrConsensusQueueNotImplemented   = whoops.Errorf("consensus queue not implemented for queueTypeName %s")
	ErrUnableToFindPubKeyForValidator = whoops.Errorf("unable to find public key for validator: %s")
	ErrSignatureVerificationFailed    = whoops.Errorf("signature verification failed (msgId: %d, valAddr: %s, pubKey: %s)")
	ErrInvalidLimitValue              = whoops.Errorf("invalid N value: %d")
)
