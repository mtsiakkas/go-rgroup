package rgroup

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LoggerData - struct containing info about the handled request.
// Passed to the postprocessor.
type LoggerData struct {
	Timestamp    int64
	ResponseSize int
	Error        *HandlerError
	Request      http.Request
	Response     *HandlerResponse
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

// Message returns the log message of the request
func (r *LoggerData) Message() string {
	if r.Error != nil {
		return r.Error.Error()
	}

	if r.Response != nil {
		return r.Response.LogMessage
	}

	return ""
}

// Status returns the resulting http status sent to the client
func (r *LoggerData) Status() int {
	if r.Error != nil {
		return r.Error.HTTPStatus
	}

	if r.Response != nil {
		return r.Response.HTTPStatus
	}

	return http.StatusOK
}

// Path returns the base uri of the request
func (r *LoggerData) Path() string {
	return strings.Split(r.Request.RequestURI, "?")[0]
}

// Duration - calculate request duration
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
		return fmt.Sprintf("%s %d %s [%3.1f%s]\n%s", r.Request.Method, r.Status(), r.Path(), dur, units, r.Message())
	}

	return fmt.Sprintf("%s %d %s [%3.1f%s]", r.Request.Method, r.Status(), r.Path(), dur, units)
}
