package rgroup

import (
	"fmt"
	"net/http"
)

// Error struct that can be used to return additional info on Handler error
type HandlerError struct {
	err        error
	LogMessage string
	Response   string
	HTTPStatus int
	Headers    http.Header
}

// Create new HandlerError with the specified http status code.
func Error(code int) *HandlerError {
	e := HandlerError{
		HTTPStatus: code,
		err:        nil,
		LogMessage: "",
		Response:   "",
		Headers:    http.Header{},
	}

	return &e
}

// Add a log message to the HandlerError.
// This message is not sent to the client.
func (e *HandlerError) WithMessage(message string, args ...any) *HandlerError {
	e.LogMessage = fmt.Sprintf(message, args...)

	return e
}

// Add response to the HandlerError to be send to the client.
func (e *HandlerError) WithResponse(response string, args ...any) *HandlerError {
	e.Response = fmt.Sprintf(response, args...)

	return e
}

// Set a response header, replacing any existing values for the header.
func (e *HandlerError) WithHeader(header string, value string) *HandlerError {
	if e.Headers == nil {
		e.Headers = http.Header{}
	}

	e.Headers.Set(header, value)

	return e
}

// Append a value to a response header, keeping any existing values.
func (e *HandlerError) AddHeader(header string, value string) *HandlerError {
	if e.Headers == nil {
		e.Headers = http.Header{}
	}

	e.Headers.Add(header, value)

	return e
}

// Delete all values of a response header.
func (e *HandlerError) DeleteHeader(header string) *HandlerError {
	e.Headers.Del(header)

	return e
}

func (e *HandlerError) Error() string {
	if e.err != nil {
		if e.LogMessage != "" {
			return fmt.Sprintf("%s: %s", e.LogMessage, e.err)
		}

		return e.err.Error()
	}

	return e.LogMessage
}

func (e *HandlerError) Wrap(err error) *HandlerError {
	e.err = err

	return e
}

func (e *HandlerError) Unwrap() error {
	return e.err
}

// Create Envelope from error.
func (e *HandlerError) ToEnvelope() *Envelope {
	env := Envelope{
		Data: nil,
		Status: EnvelopeStatus{
			HTTPStatus: e.HTTPStatus,
			Message:    nil,
			Error:      nil,
		},
	}

	if e.Response != "" {
		env.Status.Error = &e.Response
	} else {
		statusText := http.StatusText(e.HTTPStatus)
		if statusText != "" {
			env.Status.Error = &statusText
		} else {
			env.Status.Error = toPtr("unkown error")
		}
	}

	if Config.Envelope.forwardLogMessage {
		env.Status.Message = toPtr(e.Error())
	}

	return &env
}

// PanicError carries a value recovered from a panic in a Handler, along with the stack trace captured at the point of recovery.
type PanicError struct {
	value any
	stack []byte
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %v\n%s", p.value, p.stack)
}

// Value returns the value the Handler panicked with.
func (p *PanicError) Value() any {
	return p.value
}

// Stack returns the stack trace captured when the panic was recovered.
func (p *PanicError) Stack() []byte {
	return p.stack
}
