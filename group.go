package rgroup

import (
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
)

// Middleware function signature
type Middleware func(Handler) Handler

// HandlerMap is a wrapper around map[string]Handler.
// Used to simplify HandlerGroup initialization.
type HandlerMap map[string]Handler

// HandlerGroup contains all Handlers, Middleware and the custom logger for a route.
// All setup (handlers, middleware, logger) should be completed before attempting to serve from the group.
type HandlerGroup struct {
	h          http.HandlerFunc
	handlers   HandlerMap
	raw        HandlerMap
	logger     func(*LoggerData)
	middleware []Middleware
	allow      []string
	frozen     atomic.Bool
}

// MethodsAllowed returns a comma-separated string with all http verbs handled by the group
func (h *HandlerGroup) MethodsAllowed() string {
	return strings.Join(h.allow, ",")
}

func (h *HandlerGroup) methodsAllowed() {
	opts := make([]string, 1)
	opts[0] = http.MethodOptions

	for k := range h.raw {
		if k == http.MethodOptions {
			continue
		}
		opts = append(opts, k)
	}
	sort.Strings(opts)
	h.allow = opts
}

func (h *HandlerGroup) checkFrozen() {
	if h.frozen.Load() {
		panic("[rgroup] HandlerGroup build after serve")
	}
}

func (h *HandlerGroup) initMaps() {
	if h.handlers == nil {
		h.handlers = make(HandlerMap)
	}

	if h.raw == nil {
		h.raw = make(HandlerMap)
	}
}

// Create a new empty handler group
func New() *HandlerGroup {
	h := new(HandlerGroup)
	h.initMaps()
	h.build()

	return h
}

// Creates a new HandlerGroup from a HandlerMap.
func NewWithHandlers(handlers HandlerMap) *HandlerGroup {
	h := new(HandlerGroup)

	h.initMaps()
	for k, f := range handlers {
		h.raw[strings.ToUpper(k)] = f
	}

	h.build()
	return h
}

// Set a local logger function to the HandlerGroup.
// This will replace the global logger for the specified route.
func (h *HandlerGroup) SetLogger(p func(*LoggerData)) {
	h.initMaps()
	h.logger = p
	h.build()
}

// Adds a new Handler to the HandlerGroup.
func (h *HandlerGroup) AddHandler(method string, handler Handler) {
	h.initMaps()
	h.raw[strings.ToUpper(method)] = handler
	h.build()
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
func (h *HandlerGroup) AddMiddleware(m ...Middleware) *HandlerGroup {
	h.initMaps()
	if h.middleware == nil {
		h.middleware = make([]Middleware, 0)
	}

	h.middleware = append(h.middleware, m...)

	h.build()

	return h
}

func (h *HandlerGroup) build() {
	ensureLocked()
	h.checkFrozen()
	h.methodsAllowed()

	logger := config.logger
	if h.logger != nil {
		logger = h.logger
	}

	for k, f := range h.raw {
		h.handlers[k] = f.applyMiddleware(h.middleware)
	}

	h.h = func(w http.ResponseWriter, req *http.Request) {
		l := fromRequest(*req)

		func() {
			defer recoverPanic(l)

			f, ok := h.handlers[req.Method]

			switch {
			case ok:
				l.Response, l.err = f(w, req)
			case req.Method == http.MethodOptions:
				l.Response = Response(nil).WithHeader("Allow", h.MethodsAllowed())
			default:
				l.err = Error(http.StatusMethodNotAllowed).WithHeader("Allow", h.MethodsAllowed())
			}
		}()

		logAndWrite(w, l, logger)
	}
}

func (h *HandlerGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ensureLocked()
	h.frozen.CompareAndSwap(false, true)
	h.h(w, req)
}
