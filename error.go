package rgroup

import "fmt"

// Create new HandlerError with code
func Error(code int) *HandlerError {
	e := HandlerError{
		HttpStatus: code,
		err:        nil,
		LogMessage: "",
		Response:   "",
	}
	return &e
}

type HandlerError struct {
	err        error
	LogMessage string
	Response   string
	HttpStatus int
}

func (e *HandlerError) WithMessage(message string, args ...any) *HandlerError {
	e.LogMessage = fmt.Sprintf(message, args...)
	return e
}

func (e *HandlerError) WithResponse(response string, args ...any) *HandlerError {
	e.Response = fmt.Sprintf(response, args...)
	return e
}

func (e HandlerError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s", e.LogMessage, e.err)
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
