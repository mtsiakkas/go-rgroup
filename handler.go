package rgroup

import (
	"net/http"
)

// Handler function signuture
type Handler func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error)

type handlers interface {
	setInheritedMiddleware([]Middleware)
	http.Handler
}

func (h Handler) applyMiddleware(middleware []Middleware) Handler {
	f := h
	for _, m := range middleware {
		f = m(f)
	}

	return f
}

func (h Handler) ToHandlerFunc() http.HandlerFunc {

	logger := config.logger

	return func(w http.ResponseWriter, req *http.Request) {
		l := fromRequest(*req)

		func() {
			defer recoverPanic(l)

			l.Response, l.err = h(w, req)
		}()

		logAndWrite(w, l, logger)
	}
}
