package rgroup

import (
	"fmt"
)

// Create new HandlerError with the specified http status code.
func Error(code int) *HandlerError {
	e := HandlerError{
		HTTPStatus: code,
		err:        nil,
		LogMessage: "",
		Response:   "",
	}

	return &e
}

// Error struct that can be used to return additional info on Handler error
type HandlerError struct {
	err        error
	LogMessage string
	Response   string
	HTTPStatus int
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
	}

	if Config.envelopeResponse != nil && Config.envelopeResponse.forwardLogMessage {
		env.Status.Message = &e.LogMessage
	}

	return &env
}
