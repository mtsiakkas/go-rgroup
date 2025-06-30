package rgroup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type Handler func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error)
type Middleware func(Handler) Handler
type HandlerMap map[string]Handler
type HandlerGroup struct {
	handlers      HandlerMap
	postprocessor func(context.Context, *RequestData)
	middleware    []Middleware
}

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

// Create new empty handler group
func New() *HandlerGroup {
	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)
	return h
}

// Create a new handler group for handler map.
// If handlers contains an options key then behaviour is defined by the global OptionsHandlerBehaviour option
func NewWithHandlers(handlers HandlerMap) *HandlerGroup {
	if _, ok := handlers[http.MethodOptions]; ok {
		switch GetOnOptionsHandler() {
		case OptionsHandlerPanic:
			panic("cannot overwrite options handler")
		case OptionsHandlerOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case OptionsHandlerIgnore:
			delete(handlers, http.MethodOptions)
			fmt.Print("ignoring OPTIONS handler")
		default:
			panic(fmt.Sprintf("unknown OptionsHandlerBehaviour option %s", GetOnOptionsHandler()))
		}
	}

	h := new(HandlerGroup)
	h.handlers = make(HandlerMap)

	for k, f := range handlers {
		_ = h.AddHandler(k, f)
	}
	return h
}
func (h *HandlerGroup) SetPostprocessor(p func(context.Context, *RequestData)) {
	h.postprocessor = p
}

func (h *HandlerGroup) AddHandler(method string, handler Handler) error {
	if h.handlers == nil {
		h.handlers = make(HandlerMap)
	}

	m := strings.ToUpper(method)
	if _, ok := h.handlers[m]; ok {
		switch GetDuplicateMethod() {
		case DuplicateMethodPanic:
			panic("cannot overwrite options handler")
		case DuplicateMethodIgnore:
			fmt.Print("ignoring duplicate handler")
			return nil
		case DuplicateMethodOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case DuplicateMethodError:
			return fmt.Errorf("handler for %s already set", m)
		default:
			panic(fmt.Sprintf("unknown DuplicateMethodBehaviour option %d", GetDuplicateMethod()))
		}
	}

	h.handlers[m] = handler

	return nil
}

func (h *HandlerGroup) Post(handler Handler) error {
	return h.AddHandler("POST", handler)
}

func (h *HandlerGroup) Put(handler Handler) error {
	return h.AddHandler("PUT", handler)
}

func (h *HandlerGroup) Patch(handler Handler) error {
	return h.AddHandler("PATCH", handler)
}

func (h *HandlerGroup) Delete(handler Handler) error {
	return h.AddHandler("DELETE", handler)
}

func (h *HandlerGroup) Get(handler Handler) error {
	return h.AddHandler("GET", handler)
}

func (h Handler) ApplyMiddleware(middleware []Middleware) Handler {
	f := h
	for _, m := range middleware {
		f = m(f)
	}

	return f
}

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
		if f, ok := h.handlers[req.Method]; ok && GetOnOptionsHandler() == OptionsHandlerOverwrite {
			return f.ApplyMiddleware(h.middleware)(w, req)
		}
		w.Header().Set("Allow", strings.Join(h.MethodsAllowed(), ","))
		return nil, nil
	}

	if f, ok := h.handlers[req.Method]; ok {
		// apply middleware
		return f.ApplyMiddleware(h.middleware)(w, req)
	}
	// if method is not found in group, return MethodNotAllowed error
	return nil, Error(http.StatusMethodNotAllowed)
}

// Generate http.HandlerFunc from HandlerGroup
func (h HandlerGroup) Make() http.HandlerFunc {
	if len(h.handlers) == 0 {
		return func(w http.ResponseWriter, req *http.Request) {}
	}

	// set handler request postprocessor
	// local > global > default
	if h.postprocessor == nil {
		g := GetGlobalPostprocessor()
		if g != nil {
			h.postprocessor = g
		} else {
			h.postprocessor = print
		}
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		l := FromRequest(req)
		res, err := h.serve(w, req)
		l.Time()

		defer func() {
			if req.Method == http.MethodOptions {
				if config.PostprocessOptions {
					h.postprocessor(ctx, l)
				}
			} else {
				h.postprocessor(ctx, l)
			}
		}()

		if err != nil {
			l.IsError = true
			l.Message = err.Error()

			me := new(HandlerError)
			if errors.As(err, &me) {
				l.Status = me.HttpStatus
			} else {
				l.Status = http.StatusInternalServerError
			}
			if config.EnvelopeResponse {
				env := Envelope{
					Status: EnvelopeStatus{
						Error:    &l.Message,
						HttpCode: l.Status,
						Message:  nil,
					},
					Data: nil,
				}
				l.ResponseSize, _ = write(w, env)
				return
			}

			http.Error(w, me.Response, l.Status)
			return
		}

		if res != nil {

			if res.LogMessage != "" {
				l.Message = res.LogMessage
			}

			if http.StatusText(res.HttpStatus) != "" {
				l.Status = res.HttpStatus
			}

			if config.EnvelopeResponse && reflect.TypeFor[[]byte]() != reflect.TypeOf(res.Data) {
				env := Envelope{
					Data: res.Data,
					Status: EnvelopeStatus{
						HttpCode: l.Status,
						Message:  nil,
						Error:    nil,
					},
				}
				if config.ForwardLogMessage && l.Message != "" {
					env.Status.Message = &l.Message
				}

				if config.ForwardHttpStatus {
					if l.Status != http.StatusOK {
						w.WriteHeader(l.Status)
					}
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
}
