package liberr

import (
	"errors"
	"fmt"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func (e Error) Format(a ...any) internalErr {
	return internalErr{
		e:   e,
		msg: fmt.Sprintf(e.Error(), a...),
	}
}

func (e Error) JoinErrorf(format string, args ...any) Error {
	return e.Join(fmt.Errorf(format, args...))
}

func (e Error) Join(errs ...error) Error {
	return Error(errors.Join(append([]error{e}, errs...)...).Error())
}

type internalErr struct {
	e   Error
	msg string
}

func (e internalErr) Error() string {
	return e.msg
}

func (e internalErr) Is(target error) bool {
	return errors.Is(e.e, target)
}

func (e internalErr) JoinErrorf(format string, args ...any) internalErr {
	e.msg = errors.Join(e, fmt.Errorf(format, args...)).Error()
	return e
}
