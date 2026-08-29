package rgroup

import (
	"fmt"
	"net/http"
)

// Response creates a new HandlerResponse carrying data, with status 200 OK.
func Response(data any) *HandlerResponse {
	res := HandlerResponse{
		Data:       data,
		HTTPStatus: http.StatusOK,
		LogMessage: "",
		Headers:    http.Header{},
	}

	return &res
}

// HandlerResponse is the successful result of a Handler.
// Data is written to the client as-is if it is a string or a []byte, and JSON
// marshalled otherwise. A nil Data writes no body.
// LogMessage is passed to the logger and is not sent to the client, unless
// envelope responses are enabled with Config.Envelope.SetForwardLogMessage.
type HandlerResponse struct {
	Data       any
	HTTPStatus int
	LogMessage string
	Headers    http.Header
}

// WithHTTPStatus sets the http status code of the response.
func (r *HandlerResponse) WithHTTPStatus(code int) *HandlerResponse {
	r.HTTPStatus = code

	return r
}

// WithMessage sets the log message of the response, formatted as by
// fmt.Sprintf. This message is not sent to the client by default.
func (r *HandlerResponse) WithMessage(message string, args ...any) *HandlerResponse {
	r.LogMessage = fmt.Sprintf(message, args...)

	return r
}

// WithHeader sets a response header, replacing any existing values for it.
func (r *HandlerResponse) WithHeader(header string, value string) *HandlerResponse {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}

	r.Headers.Set(header, value)

	return r
}

// AddHeader appends a value to a response header, keeping any existing values.
func (r *HandlerResponse) AddHeader(header string, value string) *HandlerResponse {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}

	r.Headers.Add(header, value)

	return r
}

// SetCookie appends a Set-Cookie header to the response.
// A nil or invalid cookie is ignored.
func (r *HandlerResponse) SetCookie(cookie *http.Cookie) *HandlerResponse {
	if cookie == nil {
		return r
	}

	if v := cookie.String(); v != "" {
		r.AddHeader("Set-Cookie", v)
	}

	return r
}

// DeleteHeader deletes all values of a response header.
func (r *HandlerResponse) DeleteHeader(header string) *HandlerResponse {
	r.Headers.Del(header)

	return r
}

// ToEnvelope creates an Envelope from the response.
// The log message is included only if
// Config.Envelope.SetForwardLogMessage is enabled.
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

// EnvelopeStatus is the status field of an Envelope.
// Message and Error are omitted from the JSON output when unset.
type EnvelopeStatus struct {
	HTTPStatus int     `json:"http_status"`
	Message    *string `json:"message,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// Envelope is the fixed structure sent to the client when envelope responses
// are enabled with Config.Envelope.Enable.
// Data is omitted from the JSON output when the handler returns no data.
type Envelope struct {
	Data   any            `json:"data,omitempty"`
	Status EnvelopeStatus `json:"status"`
}
