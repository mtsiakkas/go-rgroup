package rgroup

import (
	"context"
	"fmt"
)

// GlobalConfig defines all global configuration options
type GlobalConfig struct {
	DuplicateMethod      DuplicateMethodBehaviour
	OptionsHandler       OptionsHandlerBehaviour
	PostprocessOptions   bool
	EnvelopeResponse     bool
	ForwardHTTPStatus    bool
	ForwardLogMessage    bool
	RequestPostProcessor func(context.Context, *RequestData)
}

var config = GlobalConfig{
	PostprocessOptions:   true,
	DuplicateMethod:      0,
	OptionsHandler:       0,
	EnvelopeResponse:     false,
	ForwardHTTPStatus:    false,
	ForwardLogMessage:    false,
	RequestPostProcessor: nil,
}

// SetGlobalConfig - Self explanatory
func SetGlobalConfig(cfg GlobalConfig) {
	config = cfg
}

// DuplicateMethodBehaviour defines what should happen if the Handler for a method is reassigned
type DuplicateMethodBehaviour int

/*
DuplicateMethodPanic - panic (default).
DuplicateMethodError - return error.
DuplicateMethodOverwrite - replace existing Handler.
DuplicateMethodIgnore - ignore new Handler keeping old.
*/
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

// Validate - ensure d is a valid DuplicateMethodBehaviour option
func (d DuplicateMethodBehaviour) Validate() bool {
	_, ok := duplicateMethodOpts[d]

	return ok
}

// Implement Stringer interface
func (d DuplicateMethodBehaviour) String() string {
	return duplicateMethodOpts[d]
}

// DuplicateMethodUknownOptionError - simple error struct returned by OnDuplicateMethod when passed option is invalid
type DuplicateMethodUknownOptionError struct {
	option DuplicateMethodBehaviour
}

func (e DuplicateMethodUknownOptionError) Error() string {
	return fmt.Sprintf("unknown DuplicateMethodBehaviour option %d", e.option)
}

// OnDuplicateMethod - defines duplicate method behaviour
// returns DuplicateMethodUknownOptionError error if invalid option is passed.
func OnDuplicateMethod(o DuplicateMethodBehaviour) error {
	if !o.Validate() {
		return DuplicateMethodUknownOptionError{option: o}
	}

	config.DuplicateMethod = o

	return nil
}

// GetDuplicateMethod - return current duplicate method setting
func GetDuplicateMethod() DuplicateMethodBehaviour {
	return config.DuplicateMethod
}

// OptionsHandlerBehaviour defines what should happen if the OPTIONS handler is manually set
type OptionsHandlerBehaviour int

/*
OptionsHandlerPanic - panic (default).
OptionsHandlerIgnore - ignore new Handler keeping old.
OptionsHandlerOverwrite - replace default handler.
*/
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

// Validate - ensure d is a valid OptionsHandlerBehaviour option
func (o OptionsHandlerBehaviour) Validate() bool {
	_, ok := optsOpts[o]

	return ok
}

// OptionsHandlerUknownOptionError - simple error struct returned by OnOptionsHandler when passed option is invalid
type OptionsHandlerUknownOptionError struct {
	option OptionsHandlerBehaviour
}

func (e OptionsHandlerUknownOptionError) Error() string {
	return fmt.Sprintf("unknown OptionsHandlerBehaviour option %d", e.option)
}

// OnOptionsHandler - set options method overwrite setting
func OnOptionsHandler(o OptionsHandlerBehaviour) error {
	if !o.Validate() {
		return OptionsHandlerUknownOptionError{option: o}
	}

	config.OptionsHandler = o

	return nil
}

// GetOnOptionsHandler - return the current options method overwrite behaviour
func GetOnOptionsHandler() OptionsHandlerBehaviour {
	return config.OptionsHandler
}

// SetGlobalPostprocessor - set global request post processor
func SetGlobalPostprocessor(p func(context.Context, *RequestData)) {
	config.RequestPostProcessor = p
}

// GetGlobalPostprocessor - get global request post processor
func GetGlobalPostprocessor() func(context.Context, *RequestData) {
	return config.RequestPostProcessor
}

// SetPostprocessOptions - self explanaroty
func SetPostprocessOptions(b bool) {
	config.PostprocessOptions = b
}

// SetForwardLogMessage - self explanatory
func SetForwardLogMessage(b bool) {
	config.ForwardLogMessage = b
}

// SetForwardHTTPStatus - self explanatory
func SetForwardHTTPStatus(b bool) {
	config.ForwardHTTPStatus = b
}

// SetEnvelopeResponse - self explanatory
func SetEnvelopeResponse(b bool) {
	config.EnvelopeResponse = b
}
