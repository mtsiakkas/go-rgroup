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
