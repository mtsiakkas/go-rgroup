package rgroup

import (
	"fmt"
	"net/http"
)

// Create new HandlerResponse with data
func Response(data any) *HandlerResponse {
	res := HandlerResponse{
		Data:       data,
		HttpStatus: http.StatusOK,
		LogMessage: "",
	}
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

type EnvelopeStatus struct {
	HttpCode int     `json:"http_status"`
	Message  *string `json:"message,omitempty"`
	Error    *string `json:"error,omitempty"`
}

type Envelope struct {
	Data   any            `json:"data,omitempty"`
	Status EnvelopeStatus `json:"status"`
}
