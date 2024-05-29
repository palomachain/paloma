package consensus

import (
	"github.com/VolumeFi/whoops"
)

const (
	ErrIncorrectMessageType         = whoops.Errorf("underlying message type does not match: %T")
	ErrUnableToSaveMessageWithoutID = whoops.String("unable to save message without an ID")
	ErrMessageDoesNotExist          = whoops.Errorf("message id %d does not exist")
	ErrInvalidSignature             = whoops.String("signature is invalid")
	ErrAlreadySignedWithKey         = whoops.Errorf("message %d (%s) already signed with the key: %s")
	ErrValidatorAlreadySigned       = whoops.Errorf("validator already signed: %s")

	ErrAttestorNotSetForMessage = whoops.Errorf("attestator must be set for message: %T")

	ErrNilTypeCheck             = whoops.String("TypeCheck can't be nil")
	ErrNilBytesToSignCalculator = whoops.String("BytesToSignCalculator can't be nil")
	ErrNilVerifySignature       = whoops.String("VerifySignature can't be nil")
	ErrEmptyChainType           = whoops.String("chain type can't be empty")
	ErrEmptyChainReferenceID    = whoops.String("chain id can't be empty")
	ErrEmptyQueueTypeName       = whoops.String("queue type name can't be empty")
)
