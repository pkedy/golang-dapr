package errorz

import (
	"fmt"
)

type Error struct {
	Type     string      `json:"type"`
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Details  interface{} `json:"details,omitempty"`
	Metadata Metadata    `json:"-"`
	Err      error       `json:"-"`
}

type Metadata map[string]interface{}

func Internal(err error, format string, args ...interface{}) *Error {
	message := fmt.Sprintf(format, args...)
	return Build("INTERNAL_SERVER_ERROR", 500, message).
		Metadata(Metadata{"error": err}).
		Error(err).
		Err()
}

func NotFound(format string, args ...interface{}) *Error {
	message := fmt.Sprintf(format, args...)
	return New("NOT_FOUND", 404, message)
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	if errz, ok := err.(*Error); ok {
		return errz
	}
	return Internal(err, err.Error())
}

func New(t string, code int, message string, details ...interface{}) *Error {
	var detailsItem interface{}
	if len(details) > 0 {
		detailsItem = details[0]
	}
	return &Error{
		Type:    t,
		Code:    code,
		Message: message,
		Details: detailsItem,
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) WithMessage(format string, args ...interface{}) *Error {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

func (e *Error) WithDetails(details interface{}) *Error {
	e.Details = details
	return e
}

func (e *Error) WithMetadata(metadata Metadata) *Error {
	e.Metadata = metadata
	return e
}

func (e *Error) WithError(err error) *Error {
	e.Err = err
	return e
}

type Builder struct {
	err *Error
}

func Build(t string, code int, message ...string) Builder {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}
	return Builder{
		err: New(t, code, msg),
	}
}

func (b Builder) Message(message string) Builder {
	b.err.Message = message
	return b
}

func (b Builder) Messagef(format string, args ...interface{}) Builder {
	b.err.Message = fmt.Sprintf(format, args...)
	return b
}

func (b Builder) Details(details interface{}) Builder {
	b.err.Details = details
	return b
}

func (b Builder) Metadata(metadata Metadata) Builder {
	b.err.Metadata = metadata
	return b
}

func (b Builder) Error(err error) Builder {
	b.err.Err = err
	return b
}

func (b Builder) Err() *Error {
	return b.err
}
