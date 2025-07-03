package rgroup

import (
	"fmt"
	"net/http"
)

// Response - Create new HandlerResponse with data
func Response(data any) *HandlerResponse {
	res := HandlerResponse{
		Data:       data,
		HTTPStatus: http.StatusOK,
		LogMessage: "",
	}

	return &res
}

// HandlerResponse - Handler return type on success
type HandlerResponse struct {
	Data       any
	HTTPStatus int
	LogMessage string
}

// WithHTTPStatus - set HTTP status code
func (r *HandlerResponse) WithHTTPStatus(code int) *HandlerResponse {
	r.HTTPStatus = code

	return r
}

// WithMessage - set log message
func (r *HandlerResponse) WithMessage(message string, args ...any) *HandlerResponse {
	r.LogMessage = fmt.Sprintf(message, args...)

	return r
}

// ToEnvelope - create Envelope from response.
// Used when config.EnvelopeResponse is set.
func (r *HandlerResponse) ToEnvelope() *Envelope {
	e := Envelope{
		Data: r.Data,
		Status: EnvelopeStatus{
			HTTPStatus: r.HTTPStatus,
			Message:    nil,
			Error:      nil,
		},
	}

	if Config.forwardLogMessage && r.LogMessage != "" {
		e.Status.Message = &r.LogMessage
	}

	return &e
}

// EnvelopeStatus - status type for Envelope
type EnvelopeStatus struct {
	HTTPStatus int     `json:"http_status"`
	Message    *string `json:"message,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// Envelope - client response struct when config.EnvelopeResponse is set
type Envelope struct {
	Data   any            `json:"data,omitempty"`
	Status EnvelopeStatus `json:"status"`
}
