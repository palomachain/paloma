package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrIncorrectMessageType           = whoops.Errorf("underlying message type does not match: %T")
	ErrUnableToSaveMessageWithoutID   = whoops.String("unable to save message without an ID")
	ErrConsensusQueueNotImplemented   = whoops.Errorf("consensus queue not implemented for queueTypeName %s")
	ErrMessageDoesNotExist            = whoops.Errorf("message id %d does not exist")
	ErrUnableToFindPubKeyForValidator = whoops.Errorf("unable to find public key for validator: %s")
	ErrSignatureVerificationFailed    = whoops.Errorf("signature verification failed (msgId: %d, valAddr: %s, pubKey: %s)")
)
