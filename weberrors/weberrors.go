// Package weberrors augments errors with an additional code, data and message for convenience while writing web applications.
package weberrors

import (
	"fmt"
	"runtime"
)

// New returns a Detail instance with the fields you provide.
func New(err error, code int, message string, data interface{}) *Detail {
	return &Detail{
		Err:     err,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Detail provides an implementation of ErrorCoder, ErrorMessager and ErrorDataer.
type Detail struct {
	Err     error       `json:"-"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Cause returns the underlying error (field Err).
// We use the name Cause from github.com/pkg/errors, since
// the <a href="https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md">Go2 Error Value Draft Proposal</a>
// (which suggests the name Unwrap) is not standardized yet.
func (d *Detail) Cause() error {
	return d.Err
}

// Error implements the error interface by returning the Err.Error().
// This corresponds to the internal message.
func (d *Detail) Error() string {
	if d.Err != nil {
		return d.Err.Error()
	}
	return d.Message
}

// ErrorCode returns the error code.
func (d *Detail) ErrorCode() int {
	return d.Code
}

// ErrorMessage returns the public error message.
func (d *Detail) ErrorMessage() string {
	return d.Message
}

// ErrorData returns the data.
func (d *Detail) ErrorData() interface{} {
	return d.Data
}

// ErrorCoder has an ErrorCode() method.
type ErrorCoder interface {
	ErrorCode() int
}

// ErrorCode unwraps the error provided and returns the result of calling ErrorCode() on the first error that has implements it.
// Errors are unwrapped by calling Cause().
func ErrorCode(err error) int {

	for err != nil {
		ec, ok := err.(ErrorCoder)
		if ok {
			return ec.ErrorCode()
		}
		c, ok := err.(Causer)
		if !ok {
			break
		}
		err = c.Cause()
	}

	return 0
}

// ErrorMessager has an ErrorMessage() method.
type ErrorMessager interface {
	ErrorMessage() string
}

// ErrorMessage unwraps the error provided and returns the result of calling ErrorMessage() on the first error that has implements it.
// Errors are unwrapped by calling Cause().
func ErrorMessage(err error) string {

	for err != nil {
		ec, ok := err.(ErrorMessager)
		if ok {
			return ec.ErrorMessage()
		}
		c, ok := err.(Causer)
		if !ok {
			break
		}
		err = c.Cause()
	}

	return ""
}

// ErrorDataer has an ErrorData() method.
type ErrorDataer interface {
	ErrorData() interface{}
}

// ErrorData unwraps the error provided and returns the result of calling ErrorData() on the first error that has implements it.
// Errors are unwrapped by calling Cause().
func ErrorData(err error) interface{} {

	for err != nil {
		ec, ok := err.(ErrorDataer)
		if ok {
			return ec.ErrorData()
		}
		c, ok := err.(Causer)
		if !ok {
			break
		}
		err = c.Cause()
	}

	return nil
}

// RootCause calls Cause() recursively until it finds an error that doesn't implement it and returns that.
// If Cause() returns nil then that will be returned.
func RootCause(err error) error {
	for err != nil {
		c, ok := err.(Causer)
		if !ok {
			break
		}
		err = c.Cause()
		if err == c.(error) { // avoid infinite loop
			break
		}
	}
	return err
}

// Causer allows you to get the underlying error that caused this one, if available.
// (It incidentally is compatibile with github.com/pkg/errors).  It is recommended
// that only errors which have a cause implement this interface.
type Causer interface {
	Cause() error
}

type errLoc struct {
	cause    error
	location string
}

// Cause returns the underlying error.
func (el *errLoc) Cause() error {
	return el.cause
}

// Error returns the causes error prefixed by location information (file and line number).
func (el *errLoc) Error() string {
	return el.location + ": " + el.cause.Error()
}

// ErrLoc wraps an error so it's Error() method will return the same text prefixed
// with the file and line number it was called from.  A nil error value will return nil.
// The returned value also has a Cause() method which will return the underlying error.
func ErrLoc(err error) error {

	if err == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)
	return &errLoc{
		cause:    err,
		location: fmt.Sprintf("%s:%v", file, line),
	}
}
