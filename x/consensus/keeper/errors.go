package keeper

import (
	"github.com/vizualni/whoops"
)

const (
	ErrIncorrectMessageType         = whoops.Errorf("underlying message type does not match: should be %T, but provided type is %T")
	ErrUnableToSaveMessageWithoutID = whoops.String("unable to save message without an ID")
	ErrConsensusQueueNotImplemented = whoops.Errorf("consensus queue not implemented for queueTypeName %s")
	ErrMessageDoesNotExist          = whoops.Errorf("message id %d does not exist")
)
