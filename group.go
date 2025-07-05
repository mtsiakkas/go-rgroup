package rgroup

import (
	"errors"
	"fmt"
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
	handlers      HandlerMap
	postprocessor func(*RequestData)
	middleware    []Middleware
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
	if _, ok := handlers[http.MethodOptions]; ok {
		switch Config.overwriteOptionsHandlerBehaviour {
		case OverwriteOptionsHandlerPanic:
			panic("cannot overwrite options handler")
		case OverwriteOptionsHandlerOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case OverwriteOptionsHandlerIgnore:
			delete(handlers, http.MethodOptions)
			fmt.Print("ignoring OPTIONS handler")
		default:
			panic(fmt.Sprintf("unknown OptionsHandlerBehaviour option %s", Config.overwriteOptionsHandlerBehaviour))
		}
	}

	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)

	for k, f := range handlers {
		_ = h.AddHandler(k, f)
	}

	return h
}

// SetPostprocessor assigns a local postprocessor function to the HandlerGroup
func (h *HandlerGroup) SetPostprocessor(p func(*RequestData)) {
	h.postprocessor = p
}

// DuplicateMethodExistsError is a simple error struct indicating that a handler for the
// specified method already exists in the HandlerGroup
type DuplicateMethodExistsError struct {
	method string
}

func (e DuplicateMethodExistsError) Error() string {
	return e.method + " handler already set"
}

// AddHandler adds a new Handler to the HandlerGroup.
// In case `method` already exists, behaviour is defined by the global config.DuplicateMethod option
func (h *HandlerGroup) AddHandler(method string, handler Handler) error {
	if h.handlers == nil {
		h.handlers = make(HandlerMap)
	}

	m := strings.ToUpper(method)
	if _, ok := h.handlers[m]; ok {
		switch Config.overwriteMethodBehaviour {
		case OverwriteMethodPanic:
			panic("cannot overwrite options handler")
		case OverwriteMethodIgnore:
			fmt.Print("ignoring duplicate handler")

			return nil
		case OverwriteMethodAllow:
			fmt.Print("overwriting OPTIONS handler")
		case OverwriteMethodError:
			return DuplicateMethodExistsError{method: m}
		default:
			panic(fmt.Sprintf("unknown DuplicateMethodBehaviour option %d", Config.overwriteMethodBehaviour))
		}
	}

	h.handlers[m] = handler

	return nil
}

// Post - utility function to add POST Handler to HandlerGroup
func (h *HandlerGroup) Post(handler Handler) error {
	return h.AddHandler("POST", handler)
}

// Put - utility function to add PUT Handler to HandlerGroup
func (h *HandlerGroup) Put(handler Handler) error {
	return h.AddHandler("PUT", handler)
}

// Patch - utility function to add PATCH Handler to HandlerGroup
func (h *HandlerGroup) Patch(handler Handler) error {
	return h.AddHandler("PATCH", handler)
}

// Delete - utility function to add DELETE Handler to HandlerGroup
func (h *HandlerGroup) Delete(handler Handler) error {
	return h.AddHandler("DELETE", handler)
}

// Get - utility function to add GET Handler to HandlerGroup
func (h *HandlerGroup) Get(handler Handler) error {
	return h.AddHandler("GET", handler)
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
	if req.Method == http.MethodOptions {
		// check if custom options handler was provided
		f, ok := h.handlers[req.Method]
		if ok && Config.GetOverwriteOptionsHandlerBehaviour() == OverwriteOptionsHandlerOverwrite {
			return f.applyMiddleware(h.middleware)(w, req)
		}
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
	if h.postprocessor == nil {
		g := Config.GetGlobalPostprocessor()
		if g != nil {
			h.postprocessor = g
		} else {
			h.postprocessor = defaultPrint
		}
	}

	return func(w http.ResponseWriter, req *http.Request) {
		l := FromRequest(req)
		res, err := h.serve(w, req)
		l.Time()

		defer func() {
			if req.Method == http.MethodOptions {
				if Config.postprocessOptions {
					h.postprocessor(l)
				}
			} else {
				h.postprocessor(l)
			}
		}()

		if err != nil {
			l.IsError = true
			l.Message = err.Error()

			me := new(HandlerError)
			if !errors.As(err, &me) {
				me.HTTPStatus = http.StatusInternalServerError
				_ = me.Wrap(err)
			}

			l.Status = me.HTTPStatus

			if Config.envelopeResponse {
				env := me.ToEnvelope()
				l.ResponseSize, _ = write(w, env)

				return
			}

			http.Error(w, me.Response, l.Status)

			return
		}

		if res == nil {
			return
		}

		l.Message = res.LogMessage

		if http.StatusText(res.HTTPStatus) != "" {
			l.Status = res.HTTPStatus
		}

		if Config.envelopeResponse && reflect.TypeFor[[]byte]() != reflect.TypeOf(res.Data) {
			env := res.ToEnvelope()

			if Config.forwardHTTPStatus && (l.Status != http.StatusOK) {
				w.WriteHeader(l.Status)
			}
			l.ResponseSize, _ = write(w, env)

			return
		}

		if l.Status != http.StatusOK {
			w.WriteHeader(l.Status)
		}
		l.ResponseSize, _ = write(w, res.Data)
	}
}
