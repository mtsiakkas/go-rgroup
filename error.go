package rgroup

import (
	"fmt"
	"net/http"
)

// HandlerError is an error carrying the http status, client response, headers
// and log message to use when a Handler fails.
// An error that is not a *HandlerError is reported as 500 Internal Server Error.
// LogMessage is passed to the logger and is not sent to the client, unless
// enabled with Config.SetForwardErrorLog or
// Config.Envelope.SetForwardLogMessage.
type HandlerError struct {
	err        error
	LogMessage string
	Response   string
	HTTPStatus int
	Headers    http.Header
}

// Error creates a new HandlerError with the specified http status code.
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

// WithMessage sets the log message of the HandlerError, formatted as by
// fmt.Sprintf. This message is not sent to the client by default.
func (e *HandlerError) WithMessage(message string, args ...any) *HandlerError {
	e.LogMessage = fmt.Sprintf(message, args...)

	return e
}

// WithResponse sets the response body sent to the client, formatted as by
// fmt.Sprintf. When it is empty, no body is written.
func (e *HandlerError) WithResponse(response string, args ...any) *HandlerError {
	e.Response = fmt.Sprintf(response, args...)

	return e
}

// WithHeader sets a response header, replacing any existing values for it.
func (e *HandlerError) WithHeader(header string, value string) *HandlerError {
	if e.Headers == nil {
		e.Headers = http.Header{}
	}

	e.Headers.Set(header, value)

	return e
}

// AddHeader appends a value to a response header, keeping any existing values.
func (e *HandlerError) AddHeader(header string, value string) *HandlerError {
	if e.Headers == nil {
		e.Headers = http.Header{}
	}

	e.Headers.Add(header, value)

	return e
}

// DeleteHeader deletes all values of a response header.
func (e *HandlerError) DeleteHeader(header string) *HandlerError {
	e.Headers.Del(header)

	return e
}

// Error implements the error interface.
// It returns the log message and the wrapped error joined by ": ", or whichever
// of the two is set.
func (e *HandlerError) Error() string {
	if e.err != nil {
		if e.LogMessage != "" {
			return fmt.Sprintf("%s: %s", e.LogMessage, e.err)
		}

		return e.err.Error()
	}

	return e.LogMessage
}

// Wrap sets the error wrapped by the HandlerError, replacing any error it
// already wraps.
func (e *HandlerError) Wrap(err error) *HandlerError {
	e.err = err

	return e
}

// Unwrap returns the wrapped error, so a HandlerError can be inspected with
// errors.Is and errors.As.
func (e *HandlerError) Unwrap() error {
	return e.err
}

// ToEnvelope creates an Envelope from the error.
// The status error field is set from the client response, falling back to the
// status text of the http status code. The log message is included only if
// Config.Envelope.SetForwardLogMessage is enabled.
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
			env.Status.Error = toPtr("unknown error")
		}
	}

	if config.envelope.forwardLogMessage {
		env.Status.Message = toPtr(e.Error())
	}

	return &env
}

// PanicError carries a value recovered from a panic in a Handler, along with the stack trace captured at the point of recovery.
type PanicError struct {
	value any
	stack []byte
}

// Error implements the error interface, reporting the panic value and the
// stack trace.
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
