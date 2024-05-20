package registry

import (
	"errors"
	"fmt"
)

type ErrServiceUnavailable struct {
	message string
}

func (err *ErrServiceUnavailable) Error() string {
	return err.message
}
func (err *ErrServiceUnavailable) Is(other error) bool {
	var errRef *ErrServiceUnavailable
	return errors.As(other, &errRef)
}

func NewErrServiceUnavailable(serviceName string) error {
	return &ErrServiceUnavailable{
		message: fmt.Sprintf("service unavailable: %s", serviceName),
	}
}
