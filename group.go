package rgroup

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
)

// Handler function signuture
type Handler func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error)

// Middleware function signature
type Middleware func(Handler) Handler

// HandlerMap is a wrapper around map[string]Handler
// Used to simplify HandlerGroup initialization
type HandlerMap map[string]Handler

// HandlerGroup is a structure that contains all Handlers, Middleware and request postprocessor for a route
type HandlerGroup struct {
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

// New creates a new empty handler group
func New() *HandlerGroup {
	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)

	return h
}

// NewWithHandlers creates a new HandlerGroup from a HandlerMap
// If handlers contains an options key then behaviour is defined by the global OptionsHandlerBehaviour option
func NewWithHandlers(handlers HandlerMap) *HandlerGroup {
	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)

	for k, f := range handlers {
		_ = h.AddHandler(k, f)
	}

	return h
}

// SetLogger assigns a local logger function to the HandlerGroup
func (h *HandlerGroup) SetLogger(p func(*LoggerData)) {
	h.logger = p
}

// AddHandler adds a new Handler to the HandlerGroup.
// In case `method` already exists, behaviour is defined by the global config.DuplicateMethod option
func (h *HandlerGroup) AddHandler(method string, handler Handler) error {
	if h.handlers == nil {
		h.handlers = make(HandlerMap)
	}

	m := strings.ToUpper(method)

	h.handlers[m] = handler

	return nil
}

// Post - utility function to add POST Handler to HandlerGroup
func (h *HandlerGroup) Post(handler Handler) error {
	return h.AddHandler(http.MethodPost, handler)
}

// Put - utility function to add PUT Handler to HandlerGroup
func (h *HandlerGroup) Put(handler Handler) error {
	return h.AddHandler(http.MethodPut, handler)
}

// Patch - utility function to add PATCH Handler to HandlerGroup
func (h *HandlerGroup) Patch(handler Handler) error {
	return h.AddHandler(http.MethodPatch, handler)
}

// Delete - utility function to add DELETE Handler to HandlerGroup
func (h *HandlerGroup) Delete(handler Handler) error {
	return h.AddHandler(http.MethodDelete, handler)
}

// Get - utility function to add GET Handler to HandlerGroup
func (h *HandlerGroup) Get(handler Handler) error {
	return h.AddHandler(http.MethodGet, handler)
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

// Make generates an http.HandlerFunc from the HandlerGroup
func (h *HandlerGroup) Make() http.HandlerFunc {
	if len(h.handlers) == 0 {
		return func(_ http.ResponseWriter, _ *http.Request) {}
	}

	// set handler request postprocessor
	// local > global > default
	if h.logger == nil {
		g := Config.GetGlobalLogger()
		if g != nil {
			h.logger = g
		} else {
			h.logger = defaultLogger
		}
	}

	return func(w http.ResponseWriter, req *http.Request) {
		l := fromRequest(*req)
		res, err := h.serve(w, req)

		if Config.prewriter != nil {
			res = Config.prewriter(req, res)
		}

		defer func() {
			l.Duration()

			if req.Method != http.MethodOptions || Config.logOptions {
				h.logger(l)
			}
		}()

		if err != nil {
			me := new(HandlerError)
			if !errors.As(err, &me) {
				me.HTTPStatus = http.StatusInternalServerError
				_ = me.Wrap(err)
			}

			l.Error = me

			if Config.envelopeResponse != nil {
				env := me.ToEnvelope()
				l.ResponseSize, _ = write(w, env)

				return
			}

			http.Error(w, me.Response, l.Status())

			return
		}

		if res == nil {
			return
		}

		l.Response = res

		if Config.envelopeResponse != nil && reflect.TypeFor[[]byte]() != reflect.TypeOf(res.Data) {
			env := res.ToEnvelope()

			if Config.envelopeResponse.forwardHTTPStatus && (l.Status() != http.StatusOK) {
				w.WriteHeader(l.Status())
			}

			l.ResponseSize, _ = write(w, env)

			return
		}

		if l.Status() != http.StatusOK {
			w.WriteHeader(l.Status())
		}

		l.ResponseSize, _ = write(w, res.Data)
	}
}
