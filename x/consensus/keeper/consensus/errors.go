package consensus

import (
	"github.com/vizualni/whoops"
)

const (
	ErrIncorrectMessageType         = whoops.Errorf("underlying message type does not match: %T")
	ErrUnableToSaveMessageWithoutID = whoops.String("unable to save message without an ID")
	ErrMessageDoesNotExist          = whoops.Errorf("message id %d does not exist")
)
