package rgroup

import (
	"context"
	"fmt"
)

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

// Global state variable
var duplicateMethodBehaviour DuplicateMethodBehaviour

// Set duplicate method behaviour
// returns error if unknown option is passed
func OnDuplicateMethod(o DuplicateMethodBehaviour) error {
	if !o.Validate() {
		return fmt.Errorf("unknown option %s", o)
	}
	duplicateMethodBehaviour = o
	return nil
}

// Return current duplicate method setting
func GetDuplicateMethod() DuplicateMethodBehaviour {
	return duplicateMethodBehaviour
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

// Global state variable
var optionsHandlerBehaviour OptionsHandlerBehaviour

// Set options method overwrite setting
func OnOptionsHandler(o OptionsHandlerBehaviour) error {
	if !o.Validate() {
		return fmt.Errorf("unknown option %d", o)
	}
	optionsHandlerBehaviour = o
	return nil
}

// Return the current options method overwrite behaviour
func GetOnOptionsHandler() OptionsHandlerBehaviour {
	return optionsHandlerBehaviour
}

// Global state variable
var globalRequestPostprocessor func(context.Context, *RequestData)

// Set global request post processor
func SetGlobalPostprocessor(p func(context.Context, *RequestData)) {
	globalRequestPostprocessor = p
}

// Get global request post processor
func GetGlobalPostprocessor() func(context.Context, *RequestData) {
	return globalRequestPostprocessor
}
