package rgroup

import (
	"net/http"
)

// Handler function signuture
type Handler func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error)

func (h Handler) applyMiddleware(middleware []Middleware) Handler {
	f := h
	for _, m := range middleware {
		f = m(f)
	}

	return f
}

func (h Handler) ToHandlerFunc() http.HandlerFunc {

	logger := Config.logger

	return func(w http.ResponseWriter, req *http.Request) {
		l := fromRequest(*req)

		l.Response, l.err = h(w, req)

		logAndWrite(w, l, logger)
	}
}
