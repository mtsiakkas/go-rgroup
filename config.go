package rgroup

import (
	"context"
	"fmt"
)

type DuplicateMethodBehaviour int

const (
	DuplicateMethodPanic DuplicateMethodBehaviour = iota
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

func (d DuplicateMethodBehaviour) String() string {
	return duplicateMethodOpts[d]
}

var duplicateMethodBehaviour DuplicateMethodBehaviour

func OnDuplicateMethod(o DuplicateMethodBehaviour) error {
	if !o.Validate() {
		return fmt.Errorf("unknown option %s", o)
	}
	duplicateMethodBehaviour = o
	return nil
}

func GetDuplicateMethod() DuplicateMethodBehaviour {
	return duplicateMethodBehaviour
}

type OptionsHandlerBehaviour int

const (
	OptionsHandlerPanic OptionsHandlerBehaviour = iota
	OptionsHandlerIgnore
	OptionsHandlerOverwrite
)

var optsOpts = map[OptionsHandlerBehaviour]string{
	OptionsHandlerPanic:     "panic",
	OptionsHandlerIgnore:    "ignore",
	OptionsHandlerOverwrite: "overwrite",
}

func (o OptionsHandlerBehaviour) String() string {
	return optsOpts[o]
}

func (o OptionsHandlerBehaviour) Validate() bool {
	_, ok := optsOpts[o]
	return ok
}

var optionsHandlerBehaviour OptionsHandlerBehaviour

func OnOptionsHandler(o OptionsHandlerBehaviour) error {
	if !o.Validate() {
		return fmt.Errorf("unknown option %d", o)
	}
	optionsHandlerBehaviour = o
	return nil
}

func GetOnOptionsHandler() OptionsHandlerBehaviour {
	return optionsHandlerBehaviour
}

var globalRequestPostprocessor func(context.Context, *RequestData)

func SetGlobalPostprocessor(p func(context.Context, *RequestData)) {
	globalRequestPostprocessor = p
}

func GetGlobalPostprocessor() func(context.Context, *RequestData) {
	return globalRequestPostprocessor
}
