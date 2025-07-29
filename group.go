package rgroup

import (
	"net/http"
	"strings"
)

// Middleware function signature
type Middleware func(Handler) Handler

// HandlerMap is a wrapper around map[string]Handler.
// Used to simplify HandlerGroup initialization.
type HandlerMap map[string]Handler

// HandlerGroup contains all Handlers, Middleware and the custom logger for a route.
type HandlerGroup struct {
	h          http.HandlerFunc
	handlers   HandlerMap
	logger     func(*LoggerData)
	middleware []Middleware
}

// MethodsAllowed returns a string slice with all http verbs handled by the group
func (h *HandlerGroup) MethodsAllowed() []string {
	opts := make([]string, len(h.handlers)+1)
	opts[0] = http.MethodOptions

	i := 1
	for k := range h.handlers {
		opts[i] = k
		i++
	}

	return opts
}

// Create a new empty handler group
func New() *HandlerGroup {
	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)

	return h
}

// Creates a new HandlerGroup from a HandlerMap.
func NewWithHandlers(handlers HandlerMap) *HandlerGroup {
	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)

	for k, f := range handlers {
		h.AddHandler(k, f)
	}

	return h
}

// Set a local logger function to the HandlerGroup.
// This will replace the global logger for the specified route.
func (h *HandlerGroup) SetLogger(p func(*LoggerData)) {
	h.logger = p
}

// Adds a new Handler to the HandlerGroup.
func (h *HandlerGroup) AddHandler(method string, handler Handler) {
	if Config.lockOnMake && h.h != nil {
		return
	}

	if h.handlers == nil {
		h.handlers = make(HandlerMap)
	}

	m := strings.ToUpper(method)

	h.handlers[m] = handler
}

// Utility function to add POST Handler to HandlerGroup
func (h *HandlerGroup) Post(handler Handler) {
	h.AddHandler(http.MethodPost, handler)
}

// Utility function to add PUT Handler to HandlerGroup
func (h *HandlerGroup) Put(handler Handler) {
	h.AddHandler(http.MethodPut, handler)
}

// Utility function to add PATCH Handler to HandlerGroup
func (h *HandlerGroup) Patch(handler Handler) {
	h.AddHandler(http.MethodPatch, handler)
}

// Utility function to add DELETE Handler to HandlerGroup
func (h *HandlerGroup) Delete(handler Handler) {
	h.AddHandler(http.MethodDelete, handler)
}

// Utility function to add GET Handler to HandlerGroup
func (h *HandlerGroup) Get(handler Handler) {
	h.AddHandler(http.MethodGet, handler)
}

// AddMiddleware appends the given Middleware to the HandlerGroup
func (h *HandlerGroup) AddMiddleware(m Middleware) *HandlerGroup {
	if Config.lockOnMake && h.h != nil {
		return h
	}

	if h.middleware == nil {
		h.middleware = make([]Middleware, 0)
	}

	h.middleware = append(h.middleware, m)

	return h
}

// Generates an http.HandlerFunc from the HandlerGroup.
func (h *HandlerGroup) Make() http.HandlerFunc {
	if Config.lockOnMake && h.h != nil {
		return h.h
	}

	if len(h.handlers) == 0 {
		h.h = func(_ http.ResponseWriter, _ *http.Request) {}
		return h.h
	}

	// set handler request postprocessor
	// local > global > default
	if h.logger == nil {
		g := Config.logger
		if g != nil {
			h.logger = g
		} else {
			h.logger = defaultLogger
		}
	}

	logger := h.logger

	h.h = func(w http.ResponseWriter, req *http.Request) {
		l := fromRequest(*req)

		f, ok := h.handlers[req.Method]
		switch {
		case ok:
			l.Response, l.err = f.applyMiddleware(h.middleware)(w, req)
		case !ok && req.Method == http.MethodOptions:
			l.Response = Response(nil).WithHeader("Allow", strings.Join(h.MethodsAllowed(), ","))
		default:
			l.err = Error(http.StatusMethodNotAllowed)
		}

		logAndWrite(w, l, logger)
	}

	return h.h
}

func (h *HandlerGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Make()(w, req)
}
