package resolver

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

type ErrInvalidPath struct {
	message string
}

func (err *ErrInvalidPath) Error() string {
	return err.message
}
func (err *ErrInvalidPath) Is(other error) bool {
	var errRef *ErrInvalidPath
	return errors.As(other, &errRef)
}

func NewErrInvalidPath(path string) error {
	return &ErrInvalidPath{
		message: fmt.Sprintf("invalid path: %s", path),
	}
}

type ErrEmptyBaseUrl struct {
	message string
}

func (err *ErrEmptyBaseUrl) Error() string {
	return err.message
}
func (err *ErrEmptyBaseUrl) Is(other error) bool {
	var errRef *ErrEmptyBaseUrl
	return errors.As(other, &errRef)
}

func NewErrEmptyBaseUrl() error {
	return &ErrEmptyBaseUrl{
		message: "empty base url",
	}
}
