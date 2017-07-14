package rrd

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNilOption is returned by NewClient if an option is nil.
	ErrNilOption = errors.New("nil option")
)

// Error represents a error returned from the rrdcached server.
type Error struct {
	Code int
	Msg  string
}

// NewError returns a new Error.
func NewError(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v (%v)", e.Msg, e.Code)
}

// InvalidResponseError is the error returned when the response data was invalid.
type InvalidResponseError struct {
	Reason string
	Data   []string
}

// NewInvalidResponseError returns a new InvalidResponseError from lines.
func NewInvalidResponseError(reason string, lines ...string) *InvalidResponseError {
	return &InvalidResponseError{Reason: reason, Data: lines}
}

func (e *InvalidResponseError) Error() string {
	return fmt.Sprintf("%v (%v)", e.Reason, strings.Join(e.Data, ", "))
}
