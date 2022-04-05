package keeper

type _err string

func (e _err) Error() string { return string(e) }

const (
	ErrIncorrectMessageType         _err = "underlying message type does not match"
	ErrUnableToSaveMessageWithoutID _err = "unable to save message without an ID"
	ErrConcensusQueueNotImplemented _err = "concensus queue not implemented"
	ErrMessageDoesNotExist          _err = "message does not exist"
)
