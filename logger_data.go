package rgroup

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type LoggerData struct {
	Timestamp    int64
	ResponseSize int
	Error        *HandlerError
	Request      http.Request
	Response     *HandlerResponse
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

// Path returns the base uri of the request.
func (r *LoggerData) Path() string {
	return strings.Split(r.Request.RequestURI, "?")[0]
}

// Duration returns the time taken to handle the request.
// This method is idempotent; the duration is calculated and stored on first call.
func (r *LoggerData) Duration() int64 {
	if !r.time {
		r.duration = time.Now().UnixNano() - r.Timestamp
		r.time = true
	}

	return r.duration
}

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
