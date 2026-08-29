package rgroup

import (
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
)

// Middleware wraps a Handler and returns the wrapped Handler.
type Middleware func(Handler) Handler

// HandlerMap maps http methods to the Handler serving them.
// It is used to simplify HandlerGroup initialization.
type HandlerMap map[string]Handler

// HandlerGroup contains all Handlers, Middleware and the custom logger for a route.
// It dispatches on the request method, answers OPTIONS with the methods it
// handles, and responds 405 Method Not Allowed to anything else.
//
// A HandlerGroup implements http.Handler and can be registered on any net/http
// router. All setup (handlers, middleware, logger) must be completed before
// the group serves its first request; the group is frozen by the first call to
// ServeHTTP, and setting it up afterwards panics.
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

// MethodsAllowed returns a comma-separated string with all http verbs handled
// by the group, as sent in the Allow header. OPTIONS is always included.
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

// New creates a new empty HandlerGroup.
// It panics if the global config has not been locked with Config.Lock.
func New() *HandlerGroup {
	h := new(HandlerGroup)
	h.init()
	h.build()

	return h
}

// NewWithHandlers creates a new HandlerGroup from a HandlerMap.
// Methods are normalized to upper case.
// It panics if the global config has not been locked with Config.Lock, or if
// any Handler in the map is nil or keyed by an empty method.
func NewWithHandlers(handlers HandlerMap) *HandlerGroup {
	h := new(HandlerGroup)
	h.init()
	for m, f := range handlers {
		h.validate(m, f)
		h.raw[strings.ToUpper(m)] = f
	}

	h.build()

	return h
}

// SetLogger sets a local logger function on the HandlerGroup, replacing the
// global logger for this route only. Passing nil restores the global logger.
// It returns the group, so calls can be chained.
func (h *HandlerGroup) SetLogger(p func(*LoggerData)) *HandlerGroup {
	h.init()
	h.logger = p
	h.build()

	return h
}

func (h *HandlerGroup) validate(method string, handler Handler) {
	if handler == nil {
		panic("[rgroup.HandlerGroup] nil Handler")
	}
	if method == "" {
		panic("[rgroup.HandlerGroup] Handler without method")
	}
	if _, ok := h.raw[strings.ToUpper(method)]; ok {
		panic("[rgroup.HandlerGroup] duplicate Handler")
	}
}

// Handle sets the Handler for method, which is normalized to upper case.
// It returns the group, so calls can be chained.
// It panics if handler is nil, method is empty, or the method already has a
// Handler.
func (h *HandlerGroup) Handle(method string, handler Handler) *HandlerGroup {
	h.init()
	h.validate(method, handler)
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

// AddMiddleware appends the given Middleware to the HandlerGroup, applying it
// to every Handler in the group, including those added later.
// The first Middleware added is the innermost, closest to the Handler; any
// middleware inherited from a parent HandlerMux wraps all of it.
// It returns the group, so calls can be chained, and panics if any Middleware
// is nil.
func (h *HandlerGroup) AddMiddleware(middleware ...Middleware) *HandlerGroup {
	h.init()
	for _, m := range middleware {
		if m == nil {
			panic("[rgroup.HandlerGroup] nil middleware")
		}
	}
	h.middleware = append(h.middleware, middleware...)
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

// ServeHTTP implements http.Handler.
// The first call freezes the group; mutating it afterwards panics.
func (h *HandlerGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ensureLocked()
	h.frozen.CompareAndSwap(false, true)
	h.h(w, req)
}
