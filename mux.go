package rgroup

import (
	"net/http"
	"sync"
)

// HandlerMux is safe for concurrent use.
type HandlerMux struct {
	mtx        sync.Mutex
	built      http.Handler
	handlers   map[string]http.Handler
	middleware []Middleware
	prefix     string
}

// Create a new empty HandlerMux
func NewServeMux() *HandlerMux {
	h := new(HandlerMux)
	h.handlers = make(map[string]http.Handler)
	h.middleware = make([]Middleware, 0)

	return h
}

func (m *HandlerMux) SetPrefix(prefix string) *HandlerMux {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.prefix = prefix
	return m
}

// Add HandlerGroup
func (m *HandlerMux) Handle(path string, h http.Handler) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.handlers[path] = h
}

// Add middleware to all handler groups in mux
func (m *HandlerMux) AddMiddleware(mid ...Middleware) *HandlerMux {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.middleware = append(m.middleware, mid...)
	return m
}

// Generates an http.ServeMux from the HandlerMux.
func (m *HandlerMux) Make() http.Handler {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.built != nil {
		return m.built
	}

	s := http.NewServeMux()

	for p, h := range m.handlers {
		var h3 http.Handler
		switch h2 := h.(type) {
		case *HandlerMux:
			h2.AddMiddleware(m.middleware...)
			h3 = h2.Make()
		case *HandlerGroup:
			h2.AddMiddleware(m.middleware...)
			h3 = h2.Make()
		default:
			h3 = fromHandler(h2).applyMiddleware(m.middleware).ToHandlerFunc()
		}
		s.Handle(p, h3)
	}

	m.built = http.StripPrefix(m.prefix, s)

	return m.built
}

func (m *HandlerMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.Make().ServeHTTP(w, req)
}
