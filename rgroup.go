package rgroup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type DuplicateMethodBehaviour int

const (
	DuplicateMethodPanic DuplicateMethodBehaviour = iota
	DuplicateMethodIgnore
	DuplicateMethodOverwrite
	DuplicateMethodError
)

type OptionsHandlerBehaviour int

const (
	OptionsHandlerPanic OptionsHandlerBehaviour = iota
	OptionsHandlerIgnore
	OptionsHandlerOverwrite
)

var globalRequestPostprocessor func(context.Context, *RequestData)
var duplicateMethodBehaviour DuplicateMethodBehaviour
var optionsHandlerBehaviour OptionsHandlerBehaviour

func SetGlobalPostprocessor(p func(context.Context, *RequestData)) {
	globalRequestPostprocessor = p
}

func OnDuplicateMethod(o DuplicateMethodBehaviour) {
	duplicateMethodBehaviour = o
}

func OnOptionsHandler(o OptionsHandlerBehaviour) {
	optionsHandlerBehaviour = o
}

// Create new HandlerResponse with data
func Response(data any) *HandlerResponse {
	res := HandlerResponse{Data: data, HttpStatus: http.StatusOK}
	return &res
}

// Create new HandlerError with code
func Error(code int) *HandlerError {
	e := HandlerError{HttpStatus: code}
	return &e
}

// Create new empty handler group
func New() *HandlerGroup {
	h := HandlerGroup{
		handlers: make(map[string]Handler),
	}
	return &h
}

// Create a new handler group for handler map.
// If handlers contains an options key then behaviour is defined by the global OptionsHandlerBehaviour option
func NewWithHandlers(handlers map[string]Handler) *HandlerGroup {
	if _, ok := handlers[http.MethodOptions]; ok {
		switch optionsHandlerBehaviour {
		case OptionsHandlerPanic:
			panic("cannot overwrite options handler")
		case OptionsHandlerOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case OptionsHandlerIgnore:
			delete(handlers, http.MethodOptions)
			fmt.Print("ignoring OPTIONS handler")
		}
	}

	h := HandlerGroup{
		handlers: make(map[string]Handler),
	}

	for k, f := range handlers {
		_ = h.AddHandler(k, f)
	}
	return &h
}

/*----------------------------------------------------------------------------*/
/*                                                                            */
/* HANDLER GROUP                                                              */
/*                                                                            */
/*----------------------------------------------------------------------------*/
type Handler func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error)
type Middleware func(Handler) Handler

type HandlerGroup struct {
	handlers      map[string]Handler
	postprocessor func(context.Context, *RequestData)
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

func (h *HandlerGroup) SetPostprocessor(p func(context.Context, *RequestData)) {
	h.postprocessor = p
}

func (h *HandlerGroup) AddHandler(method string, handler Handler) error {
	m := strings.ToUpper(method)
	if _, ok := h.handlers[m]; ok {
		switch duplicateMethodBehaviour {
		case DuplicateMethodPanic:
			panic("cannot overwrite options handler")
		case DuplicateMethodIgnore:
			fmt.Print("ignoring dupliacte handler")
			return nil
		case DuplicateMethodOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case DuplicateMethodError:
			return fmt.Errorf("handler for %s already set", m)
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

func (h *HandlerGroup) serve(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
	if req.Method == "OPTIONS" {
		// check if custom options handler was provided
		if f, ok := h.handlers[req.Method]; ok && optionsHandlerBehaviour == OptionsHandlerOverwrite {
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
		return nil, &HandlerError{HttpStatus: http.StatusMethodNotAllowed}
	}
}

func (h HandlerGroup) Make() http.HandlerFunc {
	if h.postprocessor == nil {
		if globalRequestPostprocessor != nil {
			h.postprocessor = globalRequestPostprocessor
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
			h.postprocessor(ctx, l)
		}()

		if err != nil {
			me := new(HandlerError)
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
				w.WriteHeader(l.Status)
			}

			if res.Data != nil {
				switch reflect.TypeOf(res.Data) {
				case reflect.TypeFor[string]():
					l.ResponseSize, _ = w.Write([]byte(res.Data.(string)))
				case reflect.TypeFor[[]byte]():
					l.ResponseSize, _ = w.Write(res.Data.([]uint8))
				default:
					l.ResponseSize, _ = marshalAndWrite(w, res.Data)
				}
			}
			if res.LogMessage != "" {
				l.Message = res.LogMessage
			}
		}
	}
}

/*----------------------------------------------------------------------------*/
/*                                                                            */
/* HELPERS                                                                    */
/*                                                                            */
/*----------------------------------------------------------------------------*/
func print(ctx context.Context, r *RequestData) {
	printFunc := log.Printf
	if r.IsError {
		printFunc = func(s string, args ...any) { log.Printf("\033[31m"+s+"\033[0m", args...) }
	}

	dur := float32(r.Duration)
	i := 0
	units := []string{"ns", "us", "ms", "s"}
	for dur > 1000 && i < 3 {
		dur /= 1000
		i++
	}

	if r.Message != "" {
		printFunc("%s %d %s [%3.1f%s]\n%s", r.Method, r.Status, r.Path, dur, units[i], r.Message)
	} else {
		printFunc("%s %d %s [%3.1f%s]", r.Method, r.Status, r.Path, dur, units[i])
	}
}

func marshalAndWrite(w http.ResponseWriter, d any) (int, error) {
	dj, err := json.Marshal(d)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(dj)
	if err != nil {
		return 0, err
	}
	return n, nil
}

/*----------------------------------------------------------------------------*/
/*                                                                            */
/* REQUEST                                                                    */
/*                                                                            */
/*----------------------------------------------------------------------------*/
type RequestData struct {
	Id           int
	Path         string
	Params       url.Values
	Ts           int64
	Method       string
	Duration     int64
	Message      string
	Status       int
	IsError      bool
	ResponseSize int
	Context      context.Context
}

func FromRequest(req *http.Request) *RequestData {
	l := RequestData{
		Path:    strings.Split(req.RequestURI, "?")[0],
		Method:  req.Method,
		Params:  req.URL.Query(),
		Status:  http.StatusOK,
		Ts:      time.Now().UnixNano(),
		Context: req.Context(),
	}

	return &l
}

func (l *RequestData) Time() int64 {
	if l.Duration == 0 {
		l.Duration = time.Now().UnixNano() - l.Ts
	}
	return l.Duration
}

/*----------------------------------------------------------------------------*/
/*                                                                            */
/* ERROR                                                                      */
/*                                                                            */
/*----------------------------------------------------------------------------*/
type HandlerError struct {
	LogMessage string
	Response   string
	ErrorCode  string
	HttpStatus int
}

func (e *HandlerError) WithMessage(message string, args ...any) *HandlerError {
	e.LogMessage = fmt.Sprintf(message, args...)
	return e
}

func (e *HandlerError) WithResponse(response string, args ...any) *HandlerError {
	e.Response = fmt.Sprintf(response, args...)
	return e
}

func (e *HandlerError) WithCode(code string) *HandlerError {
	e.ErrorCode = code
	return e
}

func (e HandlerError) Error() string {
	return e.LogMessage
}

/*----------------------------------------------------------------------------*/
/*                                                                            */
/* RESPONSE                                                                   */
/*                                                                            */
/*----------------------------------------------------------------------------*/
type HandlerResponse struct {
	Data       any
	HttpStatus int
	LogMessage string
}

func (r *HandlerResponse) WithHttpStatus(code int) *HandlerResponse {
	r.HttpStatus = code
	return r
}

func (r *HandlerResponse) WithMessage(message string, args ...any) *HandlerResponse {
	r.LogMessage = fmt.Sprintf(message, args...)
	return r
}
