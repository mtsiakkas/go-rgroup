package rgroup

import (
	"fmt"
	"net/http"

	"github.com/mtsiakkas/go-rgroup/pkg/config"
	"github.com/mtsiakkas/go-rgroup/pkg/group"
	"github.com/mtsiakkas/go-rgroup/pkg/herror"
	"github.com/mtsiakkas/go-rgroup/pkg/response"
)

// Create new HandlerResponse with data
func Response(data any) *response.HandlerResponse {
	res := response.HandlerResponse{Data: data, HttpStatus: http.StatusOK}
	return &res
}

// Create new HandlerError with code
func Error(code int) *herror.HandlerError {
	e := herror.HandlerError{HttpStatus: code}
	return &e
}

// Create new empty handler group
func New() *group.HandlerGroup {
	return new(group.HandlerGroup)
}

// Create a new handler group for handler map.
// If handlers contains an options key then behaviour is defined by the global OptionsHandlerBehaviour option
func NewWithHandlers(handlers map[string]group.Handler) *group.HandlerGroup {
	if _, ok := handlers[http.MethodOptions]; ok {
		switch config.GetOnOptionsHandler() {
		case config.OptionsHandlerPanic:
			panic("cannot overwrite options handler")
		case config.OptionsHandlerOverwrite:
			fmt.Print("overwriting OPTIONS handler")
		case config.OptionsHandlerIgnore:
			delete(handlers, http.MethodOptions)
			fmt.Print("ignoring OPTIONS handler")
		default:
			panic(fmt.Sprintf("unknown OptionsHandlerBehaviour option %s", config.GetOnOptionsHandler()))
		}
	}

	h := new(group.HandlerGroup)

	for k, f := range handlers {
		_ = h.AddHandler(k, f)
	}
	return h
}
