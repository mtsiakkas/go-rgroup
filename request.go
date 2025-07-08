package rgroup

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RequestData - struct containing info about the handled request.
// Passed to the postprocessor.
type RequestData struct {
	Path         string
	Ts           int64
	Message      string
	Status       int
	IsError      bool
	ResponseSize int
	Request      *http.Request
	time         bool
	duration     int64
}

// FromRequest - generate RequestData struct from http.Request
func FromRequest(req *http.Request) *RequestData {
	r := RequestData{
		Path:         strings.Split(req.RequestURI, "?")[0],
		Status:       http.StatusOK,
		Ts:           time.Now().UnixNano(),
		Request:      req,
		Message:      "",
		IsError:      false,
		ResponseSize: 0,
		time:         false,
		duration:     0,
	}

	return &r
}

// Duration - calculate request duration
func (r *RequestData) Duration() int64 {
	if !r.time {
		r.duration = time.Now().UnixNano() - r.Ts
		r.time = true
	}

	return r.duration
}

func (r *RequestData) String() string {
	dur := float32(r.Duration())
	i := 0
	units := []string{"ns", "us", "ms", "s"}

	for dur > 1000 && i < 3 {
		dur /= 1000
		i++
	}

	if r.Message != "" {
		return fmt.Sprintf("%s %d %s [%3.1f%s]\n%s", r.Request.Method, r.Status, r.Path, dur, units, r.Message)
	}

	return fmt.Sprintf("%s %d %s [%3.1f%s]", r.Request.Method, r.Status, r.Path, dur, units)
}
