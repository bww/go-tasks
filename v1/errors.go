package tasks

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupported       = errors.New("Unsupported task UTD")
	ErrMalformed         = errors.New("Malformed task UTD")
	ErrInvalidParameters = errors.New("Invalid parameters")
	ErrInvalidRequest    = errors.New("Invalid request")
)

func NewServiceUnavailableError(f string) error {
	return NewRecoverable(errors.New(f))
}

type Recoverable struct {
	Cause error
}

func NewRecoverable(err error) *Recoverable {
	return &Recoverable{err}
}

func (e *Recoverable) Unwrap() error {
	return e.Cause
}

func (e *Recoverable) Error() string {
	if e.Cause == nil {
		return "Recoverable error"
	} else {
		return fmt.Sprintf("%s (recoverable)", e.Cause.Error())
	}
}
