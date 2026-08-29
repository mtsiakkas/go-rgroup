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
	inherited  []Middleware
	allow      []string
	frozen     atomic.Bool
}

var _ handlers = (*HandlerGroup)(nil)

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
		panic("[rgroup.HandlerGroup] build after serve")
	}
}

func (h *HandlerGroup) init() {
	if h.handlers == nil {
		h.handlers = make(HandlerMap)
	}

	if h.raw == nil {
		h.raw = make(HandlerMap)
	}

	if h.middleware == nil {
		h.middleware = make([]Middleware, 0)
	}

	if h.inherited == nil {
		h.inherited = make([]Middleware, 0)
	}
}

// Create a new empty handler group
func New() *HandlerGroup {
	h := new(HandlerGroup)
	h.init()
	h.build()

	return h
}

// Creates a new HandlerGroup from a HandlerMap.
func NewWithHandlers(handlers HandlerMap) *HandlerGroup {
	h := new(HandlerGroup)

	h.init()
	for m, f := range handlers {
		if f == nil {
			panic("[rgroup] attempt to add nil Handler to HandlerGroup")
		}
		if m == "" {
			panic("[rgroup] attempt to add Handler without method")
		}
		h.raw[strings.ToUpper(m)] = f
	}

	h.build()
	return h
}

// Set a local logger function to the HandlerGroup.
// This will replace the global logger for the specified route.
func (h *HandlerGroup) SetLogger(p func(*LoggerData)) *HandlerGroup {
	h.init()
	h.logger = p
	h.build()

	return h
}

// Set Handler for method.
// The method is normalized to upper case.
// Panics if method is already set
func (h *HandlerGroup) Handle(method string, handler Handler) *HandlerGroup {
	if handler == nil {
		panic("[rgroup.HandlerGroup] nil Handler")
	}
	if method == "" {
		panic("[rgroup.HandlerGroup] Handler without method")
	}
	h.init()
	if _, ok := h.raw[strings.ToUpper(method)]; ok {
		panic("[rgroup.HandlerGroup] duplicate Handler")
	}
	h.raw[strings.ToUpper(method)] = handler
	h.build()

	return h
}

func (h *HandlerGroup) setInheritedMiddleware(middleware []Middleware) {
	ms := make([]Middleware, len(middleware))
	for i, mm := range middleware {
		ms[i] = mm
	}
	h.inherited = ms
	h.build()
}

// AddMiddleware appends the given Middleware to the HandlerGroup
func (h *HandlerGroup) AddMiddleware(m ...Middleware) *HandlerGroup {
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
		h.handlers[k] = f.applyMiddleware(h.middleware).applyMiddleware(h.inherited)
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
