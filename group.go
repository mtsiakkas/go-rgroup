package rgroup

import (
	"errors"
	"net/http"
	"strings"
)

// Handler function signuture
type Handler func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error)

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

func (h Handler) applyMiddleware(middleware []Middleware) Handler {
	f := h
	for _, m := range middleware {
		f = m(f)
	}

	return f
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

func (h *HandlerGroup) serve(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
	if _, ok := h.handlers[http.MethodOptions]; !ok && req.Method == http.MethodOptions {
		w.Header().Set("Allow", strings.Join(h.MethodsAllowed(), ","))

		return Response(nil), nil
	}

	if f, ok := h.handlers[req.Method]; ok {
		// apply middleware
		return f.applyMiddleware(h.middleware)(w, req)
	}
	// if method is not found in group, return MethodNotAllowed error
	return nil, Error(http.StatusMethodNotAllowed)
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

	h.h = func(w http.ResponseWriter, req *http.Request) {
		l := fromRequest(*req)
		res, err := h.serve(w, req)

		defer func() {
			if req.Method == http.MethodOptions && !Config.logOptions {
				return
			}

			l.Duration()
			h.logger(l)
		}()

		if err != nil {
			me := new(HandlerError)
			if !errors.As(err, &me) {
				me.HTTPStatus = http.StatusInternalServerError
				_ = me.Wrap(err)
			}

			n := writeErr(w, me)

			l.Error = me
			l.ResponseSize = n

			return
		}

		if Config.prewriter != nil {
			res = Config.prewriter(req, res)
		}

		n := writeRes(w, res)

		l.Response = res
		l.ResponseSize = n
	}

	return h.h
}

func (h *HandlerGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := h.Make()
	handler(w, req)
}
