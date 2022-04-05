package keeper

type _err string

func (e _err) Error() string { return string(e) }

const (
	ErrUnableToUnpackSDKMessage     _err = "underlying message type does not match the type of the Message[T]"
	ErrUnableToSaveMessageWithoutID _err = "unable to save message without an ID"
	ErrConcensusQueueNotImplemented _err = "concensus queue not implemented"
)
