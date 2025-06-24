package rgroup

import (
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
	l := RequestData{
		Path:    strings.Split(req.RequestURI, "?")[0],
		Status:  http.StatusOK,
		Ts:      time.Now().UnixNano(),
		Request: req,
	}

	return &l
}

func (l *RequestData) Time() int64 {
	if l.Duration == 0 {
		l.Duration = time.Now().UnixNano() - l.Ts
	}
	return l.Duration
}
