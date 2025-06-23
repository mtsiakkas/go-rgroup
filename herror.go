package rgroup

import "fmt"

// Create new HandlerError with code
func Error(code int) *HandlerError {
	e := HandlerError{HttpStatus: code}
	return &e
}

type HandlerError struct {
	LogMessage string
	Response   string
	ErrorCode  string
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

func (e *HandlerError) WithCode(code string) *HandlerError {
	e.ErrorCode = code
	return e
}

func (e HandlerError) Error() string {
	return e.LogMessage
}
