package group

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mtsiakkas/go-rgroup/internal/utils"
	"github.com/mtsiakkas/go-rgroup/pkg/config"
	"github.com/mtsiakkas/go-rgroup/pkg/herror"
	"github.com/mtsiakkas/go-rgroup/pkg/request"
	"github.com/mtsiakkas/go-rgroup/pkg/response"
)

type Handler func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error)
type Middleware func(Handler) Handler

type HandlerGroup struct {
	handlers      map[string]Handler
	postprocessor func(context.Context, *request.RequestData)
}

func (h *HandlerGroup) MethodsAllowed() []string {
	opts := make([]string, len(h.handlers)+1)
	opts[0] = "OPTIONS"

	i := 1
	for k := range h.handlers {
		opts[i] = k
		i++
	}

	return opts
}

func (h *HandlerGroup) SetPostprocessor(p func(context.Context, *request.RequestData)) {
	h.postprocessor = p
}

func (h *HandlerGroup) AddHandler(method string, handler Handler) error {
	if h.handlers == nil {
		h.handlers = make(map[string]Handler)
	}

	m := strings.ToUpper(method)
	if _, ok := h.handlers[m]; ok {
		switch config.GetDuplicateMethod() {
		case config.DuplicateMethodPanic:
			panic("cannot overwrite options handler")
		case config.DuplicateMethodIgnore:
			fmt.Print("ignoring duplicate handler")
			return nil
		case config.DuplicateMethodOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case config.DuplicateMethodError:
			return fmt.Errorf("handler for %s already set", m)
		default:
			panic(fmt.Sprintf("unknown DuplicateMethodBehaviour option %d", config.GetDuplicateMethod()))
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

func (h *HandlerGroup) serve(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
	if req.Method == "OPTIONS" {
		// check if custom options handler was provided
		if f, ok := h.handlers[req.Method]; ok && config.GetOnOptionsHandler() == config.OptionsHandlerOverwrite {
			return f(w, req)
		}
		w.Header().Set("Allow", strings.Join(h.MethodsAllowed(), ","))
		return nil, nil
	}

	if f, ok := h.handlers[req.Method]; ok {
		// apply middleware
		return f(w, req)
	} else {
		// if method is not found in group, return MethodNotAllowed error
		return nil, &herror.HandlerError{HttpStatus: http.StatusMethodNotAllowed}
	}
}

// Generate http.HandlerFunc from HandlerGroup
func (h HandlerGroup) Make() http.HandlerFunc {
	// set handler request postprocessor
	// local > global > default
	if h.postprocessor == nil {
		if config.GetGlobalPostprocessor() != nil {
			h.postprocessor = config.GetGlobalPostprocessor()
		} else {
			h.postprocessor = utils.Print
		}
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		l := request.FromRequest(req)
		res, err := h.serve(w, req)
		l.Time()

		defer func() {
			h.postprocessor(ctx, l)
		}()

		if err != nil {
			me := new(herror.HandlerError)
			if errors.As(err, &me) {
				l.Status = me.HttpStatus
			} else {
				l.Status = http.StatusInternalServerError
			}
			http.Error(w, me.Response, l.Status)
			l.IsError = true
			l.Message = err.Error()
			return
		}

		if res != nil {
			if http.StatusText(res.HttpStatus) != "" {
				l.Status = res.HttpStatus
				if l.Status != http.StatusOK {
					w.WriteHeader(l.Status)
				}
			}
			l.ResponseSize, _ = utils.Write(w, res.Data)
			if res.LogMessage != "" {
				l.Message = res.LogMessage
			}
		}
	}
}
