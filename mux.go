package rgroup

import (
	"net/http"
	"sync/atomic"
)

// HandlerMux wraps an http.ServeMux and propagates its Middleware to every
// handler registered on it. A nested HandlerGroup or HandlerMux inherits the
// parent's middleware; a plain http.Handler is adapted so the middleware
// applies to it as well.
//
// Routing follows http.ServeMux pattern matching. A HandlerMux implements
// http.Handler, so it can be nested in another HandlerMux or registered on any
// net/http router. All setup (handlers, middleware, prefix) must be completed
// before the mux serves its first request; the mux is frozen by the first call
// to ServeHTTP, and mutating it afterwards panics.
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

// NewServeMux creates a new empty HandlerMux.
// It panics if the global config has not been locked with Config.Lock.
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

// SetPrefix sets a path prefix that is stripped from the request before it is
// matched against the registered patterns, so a nested HandlerMux does not have
// to repeat the path it is mounted under in its own routes.
// It returns the mux, so calls can be chained.
func (m *HandlerMux) SetPrefix(prefix string) *HandlerMux {
	m.init()
	m.prefix = prefix
	m.build()
	return m
}

// Handle registers handler for the given path pattern.
// It returns the mux, so calls can be chained.
// It panics if handler is nil, path is empty, or the path is already
// registered.
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
	m.inherited = make([]Middleware, len(middleware))
	copy(m.inherited, middleware)
	m.build()
}

// AddMiddleware appends the given Middleware to the HandlerMux, applying it to
// every handler registered on it, including those added later.
// The first Middleware added is the innermost, closest to the handler; any
// middleware inherited from a parent HandlerMux wraps all of it.
// It returns the mux, so calls can be chained, and panics if any Middleware is
// nil.
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

// ServeHTTP implements http.Handler.
// The first call freezes the mux; mutating it afterwards panics.
func (m *HandlerMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ensureLocked()
	m.frozen.CompareAndSwap(false, true)
	m.h.ServeHTTP(w, req)
}
