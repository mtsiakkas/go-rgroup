package rgroup

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LoggerData is the record of a single handled request, passed to the logger
// function once the response has been written.
// Error is set when the Handler failed, and takes precedence over Response in
// Message and Status.
type LoggerData struct {
	Timestamp    int64            // Timestamp is the time the request was received, in Unix nanoseconds.
	ResponseSize int              // ResponseSize is the number of bytes written to the client.
	Error        *HandlerError    // Error is the error the Handler failed with, or nil.
	Request      http.Request     // Request is a copy of the handled request.
	Response     *HandlerResponse // Response is the response returned by the Handler, or nil.
	err          error
	time         bool
	duration     int64
}

func fromRequest(req http.Request) *LoggerData {
	r := LoggerData{
		Timestamp:    time.Now().UnixNano(),
		Error:        nil,
		Request:      req,
		Response:     nil,
		ResponseSize: 0,
		time:         false,
		duration:     0,
	}

	return &r
}

// Message returns the log message of the request.
// If both Error and Response are nil, it returns an empty string.
func (r *LoggerData) Message() string {
	if r.Error != nil {
		return r.Error.Error()
	}

	if r.Response != nil {
		return r.Response.LogMessage
	}

	return ""
}

// Status returns the resulting http status sent to the client.
// If both Error and Response are nil, it returns 200 OK.
func (r *LoggerData) Status() int {
	if r.Error != nil {
		return r.Error.HTTPStatus
	}

	if r.Response != nil {
		return r.Response.HTTPStatus
	}

	return http.StatusOK
}

// Path returns the request uri with the query string stripped.
func (r *LoggerData) Path() string {
	return strings.Split(r.Request.RequestURI, "?")[0]
}

// Duration returns the time taken to handle the request, in nanoseconds.
// This method is idempotent; the duration is calculated and stored on first call.
func (r *LoggerData) Duration() int64 {
	if !r.time {
		r.duration = time.Now().UnixNano() - r.Timestamp
		r.time = true
	}

	return r.duration
}

// String implements fmt.Stringer, and is the format used by the builtin logger.
// It reports the method, status, path and duration of the request, with the log
// message on a second line when there is one.
func (r *LoggerData) String() string {
	dur := float32(r.Duration())
	i := 0
	units := []string{"ns", "us", "ms", "s"}

	for dur > 1000 && i < 3 {
		dur /= 1000
		i++
	}

	if r.Message() != "" {
		return fmt.Sprintf("%s %d %s [%3.1f%s]\n%s", r.Request.Method, r.Status(), r.Path(), dur, units[i], r.Message())
	}

	return fmt.Sprintf("%s %d %s [%3.1f%s]", r.Request.Method, r.Status(), r.Path(), dur, units[i])
}
