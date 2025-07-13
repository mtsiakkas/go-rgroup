package rgroup

import (
	"fmt"
	"net/http"
)

// Create new HandlerResponse with data.
func Response(data any) *HandlerResponse {
	res := HandlerResponse{
		Data:       data,
		HTTPStatus: http.StatusOK,
		LogMessage: "",
	}

	return &res
}

type HandlerResponse struct {
	Data       any
	HTTPStatus int
	LogMessage string
}

// Set HTTP status code
func (r *HandlerResponse) WithHTTPStatus(code int) *HandlerResponse {
	r.HTTPStatus = code

	return r
}

// Set log message
func (r *HandlerResponse) WithMessage(message string, args ...any) *HandlerResponse {
	r.LogMessage = fmt.Sprintf(message, args...)

	return r
}

// Create Envelope from response.
func (r *HandlerResponse) ToEnvelope() *Envelope {
	e := Envelope{
		Data: r.Data,
		Status: EnvelopeStatus{
			HTTPStatus: r.HTTPStatus,
			Message:    nil,
			Error:      nil,
		},
	}

	if Config.envelopeResponse != nil && Config.envelopeResponse.forwardLogMessage && r.LogMessage != "" {
		e.Status.Message = &r.LogMessage
	}

	return &e
}

// Status struct for Envelope
type EnvelopeStatus struct {
	HTTPStatus int     `json:"http_status"`
	Message    *string `json:"message,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// Client response struct when config.EnvelopeResponse is set
type Envelope struct {
	Data   any            `json:"data,omitempty"`
	Status EnvelopeStatus `json:"status"`
}
