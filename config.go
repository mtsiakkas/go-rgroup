package rgroup

import (
	"context"
	"fmt"
)

type GlobalConfig struct {
	DuplicateMethod      DuplicateMethodBehaviour
	OptionsHandler       OptionsHandlerBehaviour
	PostprocessOptions   bool
	EnvelopeResponse     bool
	ForwardHttpStatus    bool
	ForwardLogMessage    bool
	RequestPostProcessor func(context.Context, *RequestData)
}

var config = GlobalConfig{
	PostprocessOptions:   true,
	DuplicateMethod:      0,
	OptionsHandler:       0,
	EnvelopeResponse:     false,
	ForwardHttpStatus:    false,
	ForwardLogMessage:    false,
	RequestPostProcessor: nil,
}

func SetGlobalConfig(cfg GlobalConfig) {
	config = cfg
}

// DuplicateMethodBehaviour defines what should happen if the handler for a method is reassigned
type DuplicateMethodBehaviour int

const (
	DuplicateMethodPanic DuplicateMethodBehaviour = iota // default
	DuplicateMethodIgnore
	DuplicateMethodOverwrite
	DuplicateMethodError
)

var duplicateMethodOpts = map[DuplicateMethodBehaviour]string{
	DuplicateMethodPanic:     "panic",
	DuplicateMethodIgnore:    "ignore",
	DuplicateMethodOverwrite: "overwrite",
	DuplicateMethodError:     "error",
}

func (d DuplicateMethodBehaviour) Validate() bool {
	_, ok := duplicateMethodOpts[d]
	return ok
}

// Implement Stringer interface
func (d DuplicateMethodBehaviour) String() string {
	return duplicateMethodOpts[d]
}

// Set duplicate method behaviour
// returns error if unknown option is passed
func OnDuplicateMethod(o DuplicateMethodBehaviour) error {
	if !o.Validate() {
		return fmt.Errorf("unknown option %s", o)
	}

	config.DuplicateMethod = o
	return nil
}

// Return current duplicate method setting
func GetDuplicateMethod() DuplicateMethodBehaviour {
	return config.DuplicateMethod
}

// OptionsHandlerBehaviour defines what should happen if the OPTIONS handler is manually set
type OptionsHandlerBehaviour int

const (
	OptionsHandlerPanic OptionsHandlerBehaviour = iota // default
	OptionsHandlerIgnore
	OptionsHandlerOverwrite
)

var optsOpts = map[OptionsHandlerBehaviour]string{
	OptionsHandlerPanic:     "panic",
	OptionsHandlerIgnore:    "ignore",
	OptionsHandlerOverwrite: "overwrite",
}

// Implement Stringer interface
func (o OptionsHandlerBehaviour) String() string {
	return optsOpts[o]
}

func (o OptionsHandlerBehaviour) Validate() bool {
	_, ok := optsOpts[o]
	return ok
}

// Set options method overwrite setting
func OnOptionsHandler(o OptionsHandlerBehaviour) error {
	if !o.Validate() {
		return fmt.Errorf("unknown option %d", o)
	}

	config.OptionsHandler = o
	return nil
}

// Return the current options method overwrite behaviour
func GetOnOptionsHandler() OptionsHandlerBehaviour {
	return config.OptionsHandler
}

// Set global request post processor
func SetGlobalPostprocessor(p func(context.Context, *RequestData)) {
	config.RequestPostProcessor = p
}

// Get global request post processor
func GetGlobalPostprocessor() func(context.Context, *RequestData) {
	return config.RequestPostProcessor
}

func SetPostprocessOptions(b bool) {
	config.PostprocessOptions = b
}

func SetForwardLogMessage(b bool) {
	config.ForwardLogMessage = b
}

func SetForwardHttpStatus(b bool) {
	config.ForwardHttpStatus = b
}

func SetEnvelopeResponse(b bool) {
	config.EnvelopeResponse = b
}
