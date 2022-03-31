package client

type _err string

func (e _err) Error() string { return string(e) }

const (
	ErrSignatureVerificationFailed              _err = "signature verification failed"
	ErrSignatureDoesNotMatchItsRespectiveSigner _err = "signature does not match its respective signer"
	ErrTooLittleOrTooManySignaturesProvided     _err = "too many or too little signatures provided"
)
