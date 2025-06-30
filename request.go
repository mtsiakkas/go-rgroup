package rgroup

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RequestData struct {
	Path         string
	Ts           int64
	Duration     int64
	Message      string
	Status       int
	IsError      bool
	ResponseSize int
	Request      *http.Request
}

func FromRequest(req *http.Request) *RequestData {
	r := RequestData{
		Path:         strings.Split(req.RequestURI, "?")[0],
		Status:       http.StatusOK,
		Ts:           time.Now().UnixNano(),
		Request:      req,
		Duration:     0,
		Message:      "",
		IsError:      false,
		ResponseSize: 0,
	}

	return &r
}

func (r *RequestData) Time() int64 {
	if r.Duration == 0 {
		r.Duration = time.Now().UnixNano() - r.Ts
	}
	return r.Duration
}

func (r *RequestData) String() string {
	dur, units := timeScale(r.Duration)
	if r.Message != "" {
		return fmt.Sprintf("%s %d %s [%3.1f%s]\n%s", r.Request.Method, r.Status, r.Path, dur, units, r.Message)
	}
	return fmt.Sprintf("%s %d %s [%3.1f%s]", r.Request.Method, r.Status, r.Path, dur, units)
}
