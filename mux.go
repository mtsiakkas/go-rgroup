package rgroup

import (
	"net/http"
)

type HandlerMux struct {
	h          map[string]http.Handler
	middleware []Middleware
	prefix     string
}

// Create a new empty HandlerMux
func NewServeMux() *HandlerMux {
	h := new(HandlerMux)
	h.h = make(map[string]http.Handler)
	h.middleware = make([]Middleware, 0)

	return h
}

func (m *HandlerMux) SetPrefix(prefix string) *HandlerMux {
	m.prefix = prefix
	return m
}

// Add HandlerGroup
func (m *HandlerMux) Handle(path string, h http.Handler) {
	m.h[path] = h
}

// Add middleware to all handler groups in mux
func (m *HandlerMux) AddMiddleware(mid ...Middleware) {
	m.middleware = append(m.middleware, mid...)
}

// Generates an http.ServeMux from the HandlerMux.
func (m *HandlerMux) Make() http.Handler {
	s := new(http.ServeMux)
	for p, h := range m.h {
		var h3 http.Handler
		switch h2 := h.(type) {
		case *HandlerMux:
			h2.AddMiddleware(m.middleware...)
			h3 = h2.Make()
		case *HandlerGroup:
			h2.AddMiddleware(m.middleware...)
			h3 = h2.Make()
		default:
			hh := fromHandler(h2)
			h3 = hh.applyMiddleware(m.middleware).ToHandlerFunc()
		}
		s.Handle(p, h3)
	}
	return http.StripPrefix(m.prefix, s)
}

func (m *HandlerMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.Make().ServeHTTP(w, req)
}
