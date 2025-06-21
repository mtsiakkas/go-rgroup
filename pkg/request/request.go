package request

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RequestData struct {
	Id           int
	Path         string
	Params       url.Values
	Ts           int64
	Method       string
	Duration     int64
	Message      string
	Status       int
	IsError      bool
	ResponseSize int
	Context      context.Context
}

func FromRequest(req *http.Request) *RequestData {
	l := RequestData{
		Path:    strings.Split(req.RequestURI, "?")[0],
		Method:  req.Method,
		Params:  req.URL.Query(),
		Status:  http.StatusOK,
		Ts:      time.Now().UnixNano(),
		Context: req.Context(),
	}

	return &l
}

func (l *RequestData) Time() int64 {
	if l.Duration == 0 {
		l.Duration = time.Now().UnixNano() - l.Ts
	}
	return l.Duration
}
