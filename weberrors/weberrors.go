/*
Package weberrors augments errors with an additional code, data and message for convenience while writing web applications.
It also provides wrapping and unwrapping where a Cause() method undicates an underlying error.

The New() method will wrap the error your provide and return a new one with the provided code, message and data.
These fields correspond to the <a href="https://www.jsonrpc.org/specification#error_object" target="_blank">JSON RPC 2.0 error</a> code, message and data fields.
They are also very useful when returning errors from REST calls.

The Causer interface specifies the method Cause() error which can be used to
retrieve an underlying error - it's cause.  Any time an error is wrapped by a method in this package
the returned error will implement Causer.

The methods ErrorCode(error) int, ErrorMessage(error) string and ErrorData(error) interface{}
provide access to each of the fields wrapped with a New() call, and will "unwrap" the error by calling Cause() as necessary
to extract the needed value.  The httpapi uses these methods to extract information when

Note that weberrors.ErrorMessage(err) and err.Error() will generally return two completely different things - the former
being the <strong>public-facing error message</strong> returned to the caller and the latter being the <strong>internal error message</strong> which is usually
either checked for and handled or logged.

	// start with this error
	var baseErr = fmt.Errorf("base error")

	// wrap it with web info
	var webErr = weberrors.New(baseErr,
		501,
		"Something went wrong",
		"additional data")

	// also specify the location
	var finalErr = weberrors.ErrLoc(webErr)

	// let's see what info we can extract from it
	fmt.Printf("Error: %v\n", finalErr)
	fmt.Printf("Error Cause: %v\n", weberrors.RootCause(finalErr))
	fmt.Printf("Error Code: %v\n", weberrors.ErrorCode(finalErr))
	fmt.Printf("Error Message: %v\n", weberrors.ErrorMessage(finalErr))
	fmt.Printf("Error Data: %v\n", weberrors.ErrorData(finalErr))

*/
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

type errPrefix struct {
	cause  error
	prefix string
}

// Cause returns the underlying error.
func (ep *errPrefix) Cause() error {
	return ep.cause
}

// Error returns the cause error prefixed by the strinct provided
func (ep *errPrefix) Error() string {
	return ep.prefix + ep.cause.Error()
}

// ErrPrefix returns an error whose Error() method return will have the specified prefix.
// The error provided must not be nil and is set as the cause for the error returned.
func ErrPrefix(prefix string, err error) error {
	return &errPrefix{
		prefix: prefix,
		cause:  err,
	}
}

type errLoc struct {
	cause    error
	location string
}

// Cause returns the underlying error.
func (el *errLoc) Cause() error {
	return el.cause
}

// Error returns the cause error prefixed by location information (file and line number).
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
