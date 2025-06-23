package rgroup

import (
	"fmt"
	"net/http"
)

// Create new HandlerResponse with data
func Response(data any) *HandlerResponse {
	res := HandlerResponse{Data: data, HttpStatus: http.StatusOK}
	return &res
}

type HandlerResponse struct {
	Data       any
	HttpStatus int
	LogMessage string
}

func (r *HandlerResponse) WithHttpStatus(code int) *HandlerResponse {
	r.HttpStatus = code
	return r
}

func (r *HandlerResponse) WithMessage(message string, args ...any) *HandlerResponse {
	r.LogMessage = fmt.Sprintf(message, args...)
	return r
}
