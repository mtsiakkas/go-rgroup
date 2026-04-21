package rgroup

import (
	"fmt"
	"net/http"
)

type HandlerMux struct {
	h          map[string]*HandlerGroup
	middleware []Middleware
	prefix     string
}

// Create a new empty HandlerMux
func NewServeMux() *HandlerMux {
	h := new(HandlerMux)
	h.h = make(map[string]*HandlerGroup)
	h.middleware = make([]Middleware, 0)

	return h
}

func (m *HandlerMux) SetPrefix(prefix string) *HandlerMux {
	m.prefix = prefix
	return m
}

// Add HandlerGroup
func (m *HandlerMux) Handle(path string, h *HandlerGroup) {
	m.h[path] = h
}

// Add middleware to all handler groups in mux
func (m *HandlerMux) AddMiddleware(mid ...Middleware) {
	m.middleware = append(m.middleware, mid...)
}

// Generates an http.ServeMux from the HandlerMux.
func (m *HandlerMux) Make() *http.ServeMux {
	s := new(http.ServeMux)
	for p, h := range m.h {
		fmt.Println(p)
		for _, mid := range m.middleware {
			h.AddMiddleware(mid)
		}
		s.Handle(p, h.Make())
	}
	return http.StripPrefix(m.prefix, s)
}

func (m *HandlerMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.Make().ServeHTTP(w, req)
}
