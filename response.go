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
		Headers:    http.Header{},
	}

	return &res
}

type HandlerResponse struct {
	Data       any
	HTTPStatus int
	LogMessage string
	Headers    http.Header
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

// Set a response header, replacing any existing values for the header.
func (r *HandlerResponse) WithHeader(header string, value string) *HandlerResponse {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}

	r.Headers.Set(header, value)

	return r
}

// Append a value to a response header, keeping any existing values.
func (r *HandlerResponse) AddHeader(header string, value string) *HandlerResponse {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}

	r.Headers.Add(header, value)

	return r
}

// Append a Set-Cookie header to the response.
func (r *HandlerResponse) SetCookie(cookie *http.Cookie) *HandlerResponse {
	if cookie == nil {
		return r
	}

	if v := cookie.String(); v != "" {
		r.AddHeader("Set-Cookie", v)
	}

	return r
}

// Delete all values of a response header.
func (r *HandlerResponse) DeleteHeader(header string) *HandlerResponse {
	r.Headers.Del(header)

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

	if config.envelope.forwardLogMessage && r.LogMessage != "" {
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
