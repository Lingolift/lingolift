package errno

import (
	"errors"
	"fmt"
)

const (
	ErrTypeSender = "sender"
	ErrTypeServer = "server"
)

// Err represents a custom error type with additional context.ml:"-"`
type Err struct {
	HTTPCode int    `json:"-"`
	ErrType  string `json:"type" xml:"Type"`
	Code     string `json:"code" xml:"Code"`
	Message  string `json:"message" xml:"Message"`
	RawErr   error  `json:"-"`
}

// New creates a new Err instance.
func New(httpCode int, errCode string, msg string, err error) Err {
	return Err{
		HTTPCode: httpCode,
		ErrType:  ErrTypeSender,
		Code:     errCode,
		Message:  msg,
		RawErr:   err,
	}
}

// WrapError modifies the raw error of an Err.
func (e *Err) WrapError(err error) *Err {
	e.RawErr = err
	return e
}

// WithFormat modifies the message of an Err using a format string and arguments.
func (e *Err) WithFormat(format string, args ...interface{}) *Err {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

// WithFmt modifies the message of an Err using a format string and arguments.
func (e *Err) WithFmt(f string) Err {
	errs := Err{
		HTTPCode: e.HTTPCode,
		ErrType:  e.ErrType,
		Code:     e.Code,
		Message:  e.Message,
	}
	errs.Message = fmt.Sprintf(errs.Message, f)

	return errs
}

// WithRawErr Used to add service raw error
func (e *Err) WithRawErr(err error) Err {
	return Err{
		HTTPCode: e.HTTPCode,
		ErrType:  e.ErrType,
		Code:     e.Code,
		Message:  e.Message,
		RawErr:   err,
	}
}

// WithFmtAndRawErr
func (e *Err) WithFmtAndRawErr(f string, err error) Err {
	errs := Err{
		HTTPCode: e.HTTPCode,
		ErrType:  e.ErrType,
		Code:     e.Code,
		Message:  e.Message,
		RawErr:   err,
	}
	errs.Message = fmt.Sprintf(errs.Message, f)

	return errs
}

// WithCodeAndMessage
func (e *Err) WithCodeAndMessage(code, msg string) Err {
	return Err{
		HTTPCode: e.HTTPCode,
		ErrType:  e.ErrType,
		Code:     code,
		Message:  msg,
		RawErr:   errors.New(msg),
	}
}
