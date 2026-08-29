package rgroup

import (
	"net/http"
	"sync/atomic"
)

// HandlerMux is similar to http.ServeMux
// All setup (handlers, middleware) should be completed before attempting to serve
type HandlerMux struct {
	h          http.Handler
	m          *http.ServeMux
	handlers   map[string]http.Handler
	middleware []Middleware
	inherited  []Middleware
	prefix     string
	frozen     atomic.Bool
}

func (m *HandlerMux) checkFrozen() {
	if m.frozen.Load() {
		panic("[rgroup.HandlerMux] build after serve")
	}
}

// Create a new empty HandlerMux
func NewServeMux() *HandlerMux {
	m := new(HandlerMux)
	m.init()
	m.build()

	return m
}

var _ handlers = (*HandlerMux)(nil)

func (m *HandlerMux) init() {
	if m.handlers == nil {
		m.handlers = make(map[string]http.Handler)
	}
	if m.middleware == nil {
		m.middleware = make([]Middleware, 0)
	}
	if m.inherited == nil {
		m.inherited = make([]Middleware, 0)
	}
}

func (m *HandlerMux) SetPrefix(prefix string) *HandlerMux {
	m.init()
	m.prefix = prefix
	m.build()
	return m
}

// Add handlers to HandlerMux
func (m *HandlerMux) Handle(path string, handler http.Handler) *HandlerMux {
	if handler == nil {
		panic("[rgroup.HandlerMux] nil Handler")
	}
	if path == "" {
		panic("[rgroup.HandlerMux] Handler without path")
	}
	m.init()
	if _, ok := m.handlers[path]; ok {
		panic("[rgroup.HandlerMux] duplicate Handler")
	}
	m.handlers[path] = handler
	m.build()
	return m
}

func (m *HandlerMux) setInheritedMiddleware(middleware []Middleware) {
	ms := make([]Middleware, len(middleware))
	for i, mm := range middleware {
		ms[i] = mm
	}
	m.inherited = ms
	m.build()
}

// Add middleware to all handlers in mux
func (m *HandlerMux) AddMiddleware(middleware ...Middleware) *HandlerMux {
	m.init()
	for _, ms := range middleware {
		if ms == nil {
			panic("[rgroup.HandlerMux] nil middleware")
		}
	}
	m.middleware = append(m.middleware, middleware...)
	m.build()
	return m
}

func (m *HandlerMux) build() {
	ensureLocked()
	m.checkFrozen()

	m.m = http.NewServeMux()

	middleware := append(m.middleware, m.inherited...)

	for p, h := range m.handlers {
		switch h2 := h.(type) {
		case handlers:
			h2.setInheritedMiddleware(middleware)
			m.m.Handle(p, h2)
		default:
			m.m.Handle(p, fromHandler(h2).applyMiddleware(middleware).ToHandlerFunc())
		}
	}

	m.h = http.StripPrefix(m.prefix, m.m)
}

func (m *HandlerMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ensureLocked()
	m.frozen.CompareAndSwap(false, true)
	m.h.ServeHTTP(w, req)
}
