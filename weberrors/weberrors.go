/*
Package weberrors augments errors with an additional code, data and message for convenience while writing web applications.
It also provides wrapping and unwrapping where a Cause() method undicates an underlying error.

The New() method will wrap the error your provide and return a new one with the provided code, message and data.
These fields correspond to the JSON RPC 2.0 error code, message and data fields.
They are also very useful when returning errors from REST calls.

The Causer interface specifies the method Cause() error which can be used to
retrieve an underlying error - it's cause.  Any time an error is wrapped by a method in this package
the returned error will implement Causer.

The methods ErrorCode(error) int, ErrorMessage(error) string and ErrorData(error) interface{}
provide access to each of the fields wrapped with a New() call, and will "unwrap" the error by calling Cause() as necessary
to extract the needed value.  The httpapi uses these methods to extract information when

Note that weberrors.ErrorMessage(err) and err.Error() will generally return two completely different things - the former
being the public-facing error message returned to the caller and the latter being the internal error message which is usually
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
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

// NewCode returns an error with the cause and code you provide.
func NewCode(err error, code int) error {
	if err == nil {
		panic("nil cause error passed to weberrors.NewCode")
	}
	return &errCode{
		Err:     err,
		Code:    code,
		Headers: setErrID(nil),
	}
}

type errCode struct {
	Err     error       `json:"-"`
	Code    int         `json:"code,omitempty"`
	Headers http.Header `json:"-"`
}

func (c *errCode) Cause() error {
	return c.Err
}

func (c *errCode) Error() string {
	if c.Err != nil {
		return c.Err.Error()
	}
	return ""
}

func (c *errCode) ErrorHeaders() http.Header {
	return c.Headers
}

// New returns a error instance with the cause error, code, message, data and headers you provide.
func New(err error, code int, message string, data interface{}, headers http.Header) error {
	if err == nil {
		panic("nil cause error passed to weberrors.New")
	}
	return &detail{
		Err:     err,
		Code:    code,
		Message: message,
		Data:    data,
		Headers: setErrID(headers),
	}
}

type detail struct {
	Err     error       `json:"-"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Headers http.Header `json:"-"`
}

// Cause returns the underlying error (field Err).
// We use the name Cause from github.com/pkg/errors, since
// the Go2 Error Value Draft Proposal, which suggests the name Unwrap, is not standardized yet
// (see https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md).
func (d *detail) Cause() error {
	return d.Err
}

// Error implements the error interface by returning the Err.Error().
// This corresponds to the internal message.
func (d *detail) Error() string {
	if d.Err != nil {
		return d.Err.Error()
	}
	return d.Message
}

// ErrorCode returns the error code.
func (d *detail) ErrorCode() int {
	return d.Code
}

// ErrorMessage returns the public error message.
func (d *detail) ErrorMessage() string {
	return d.Message
}

// ErrorData returns the data.
func (d *detail) ErrorData() interface{} {
	return d.Data
}

// ErrorHeaders returns additional headers to be set on an HTTP response
func (d *detail) ErrorHeaders() http.Header {
	return d.Headers
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

// ErrorHeaderser is the interface for ErrorHeaders.
// The name is so bad it's good.  A bit like Plan 9 from Outer Space.
type ErrorHeaderser interface {
	ErrorHeaders() http.Header
}

// ErrorHeaders unwraps the error provided and returns a composite of all of the headers found during the
// unwrapping, with higher level values overwritting lower level ones.
func ErrorHeaders(err error) http.Header {

	var ret http.Header
	addRet := func(h http.Header) {
		if len(h) == 0 {
			return
		}
		if ret == nil {
			ret = make(http.Header)
		}
		// write each key that doesn't exist in the return
		for k, v := range h {
			if _, ok := ret[k]; !ok {
				ret[k] = v
			}
		}
	}

	for err != nil {
		eh, ok := err.(ErrorHeaderser)
		if ok {
			addRet(eh.ErrorHeaders()) // use whatever headers are there
		}
		c, ok := err.(Causer)
		if !ok {
			break
		}
		err = c.Cause()
	}

	return ret
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

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

func errID() string {
	return fmt.Sprintf("%020d", rnd.Uint64())
}

func setErrID(h http.Header) http.Header {
	if h == nil {
		h = make(http.Header)
	}
	if h.Get("X-Id") == "" {
		h.Set("X-Id", errID())
	}
	return h
}
