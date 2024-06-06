package consensus

import (
	"errors"

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
)

var (
	ErrNilTypeCheck             = errors.New("TypeCheck can't be nil")
	ErrNilBytesToSignCalculator = errors.New("BytesToSignCalculator can't be nil")
	ErrNilVerifySignature       = errors.New("VerifySignature can't be nil")
	ErrEmptyChainType           = errors.New("chain type can't be empty")
	ErrEmptyChainReferenceID    = errors.New("chain id can't be empty")
	ErrEmptyQueueTypeName       = errors.New("queue type name can't be empty")
)
