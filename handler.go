package rgroup

import (
	"net/http"
)

// Handler is the rgroup handler function signature.
// Instead of writing to w, a Handler returns the response to send to the client,
// or an error. Returning a *HandlerError allows the status code, client response
// and log message to be set; any other error is reported as
// 500 Internal Server Error.
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

// ToHandlerFunc adapts the Handler to an http.HandlerFunc, so it can be
// registered on any net/http router without a surrounding HandlerGroup.
// The returned function writes the response and logs the request with the
// global logger, which is captured when ToHandlerFunc is called.
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
