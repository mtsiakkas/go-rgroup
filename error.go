package rgroup

import (
	"fmt"
)

// Error - Create new HandlerError with code
func Error(code int) *HandlerError {
	e := HandlerError{
		HTTPStatus: code,
		err:        nil,
		LogMessage: "",
		Response:   "",
	}

	return &e
}

// HandlerError - error struct that can be used to return additional info on Handler error
type HandlerError struct {
	err        error
	LogMessage string
	Response   string
	HTTPStatus int
}

// WithMessage - add log message
func (e *HandlerError) WithMessage(message string, args ...any) *HandlerError {
	e.LogMessage = fmt.Sprintf(message, args...)

	return e
}

// WithResponse - add response to be send to the client
func (e *HandlerError) WithResponse(response string, args ...any) *HandlerError {
	e.Response = fmt.Sprintf(response, args...)

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

// Wrap - wrap error
func (e *HandlerError) Wrap(err error) *HandlerError {
	e.err = err

	return e
}

func (e *HandlerError) Unwrap() error {
	return e.err
}

// ToEnvelope - create Envelope from error.
// Used when config.EnvelopeResponse is set.
func (e *HandlerError) ToEnvelope() *Envelope {
	env := Envelope{
		Data: nil,
		Status: EnvelopeStatus{
			HTTPStatus: e.HTTPStatus,
			Message:    nil,
			Error:      toPtr(e.Error()),
		},
	}

	if Config.forwardLogMessage && e.Response != "" {
		env.Status.Message = &e.Response
	}

	return &env
}
