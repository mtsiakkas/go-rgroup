package rgroup

import (
	"context"
	"fmt"
)

// GlobalConfig defines all global configuration options
type GlobalConfig struct {
	duplicateMethod      DuplicateMethodBehaviour
	optionsHandler       OptionsHandlerBehaviour
	postprocessOptions   bool
	envelopeResponse     bool
	forwardHTTPStatus    bool
	forwardLogMessage    bool
	requestPostProcessor func(context.Context, *RequestData)
}

// Config is a global instance of GlobalConfig and holds the global configuration for the package
var Config = GlobalConfig{
	postprocessOptions:   true,
	duplicateMethod:      0,
	optionsHandler:       0,
	envelopeResponse:     false,
	forwardHTTPStatus:    false,
	forwardLogMessage:    false,
	requestPostProcessor: nil,
}

// Reset the global config to the default values
func (c *GlobalConfig) Reset() *GlobalConfig {
	*c = GlobalConfig{
		postprocessOptions:   true,
		duplicateMethod:      0,
		optionsHandler:       0,
		envelopeResponse:     false,
		forwardHTTPStatus:    false,
		forwardLogMessage:    false,
		requestPostProcessor: nil,
	}

	return c
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
func (c *GlobalConfig) OnDuplicateMethod(o DuplicateMethodBehaviour) error {
	if !o.Validate() {
		return DuplicateMethodUknownOptionError{option: o}
	}

	c.duplicateMethod = o

	return nil
}

// GetDuplicateMethod - return current duplicate method setting
func (c *GlobalConfig) GetDuplicateMethod() DuplicateMethodBehaviour {
	return c.duplicateMethod
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
func (c *GlobalConfig) OnOptionsHandler(o OptionsHandlerBehaviour) error {
	if !o.Validate() {
		return OptionsHandlerUknownOptionError{option: o}
	}

	c.optionsHandler = o

	return nil
}

// GetOnOptionsHandler - return the current options method overwrite behaviour
func (c *GlobalConfig) GetOnOptionsHandler() OptionsHandlerBehaviour {
	return c.optionsHandler
}

// SetGlobalPostprocessor - set global request post processor
func (c *GlobalConfig) SetGlobalPostprocessor(p func(context.Context, *RequestData)) *GlobalConfig {
	c.requestPostProcessor = p

	return c
}

// GetGlobalPostprocessor - get global request post processor
func (c *GlobalConfig) GetGlobalPostprocessor() func(context.Context, *RequestData) {
	return c.requestPostProcessor
}

// SetPostprocessOptions - self explanaroty
func (c *GlobalConfig) SetPostprocessOptions(b bool) *GlobalConfig {
	c.postprocessOptions = b

	return c
}

// SetForwardLogMessage - self explanatory
func (c *GlobalConfig) SetForwardLogMessage(b bool) *GlobalConfig {
	c.forwardLogMessage = b

	return c
}

// SetForwardHTTPStatus - self explanatory
func (c *GlobalConfig) SetForwardHTTPStatus(b bool) *GlobalConfig {
	c.forwardHTTPStatus = b

	return c
}

// SetEnvelopeResponse - self explanatory
func (c *GlobalConfig) SetEnvelopeResponse(b bool) *GlobalConfig {
	c.envelopeResponse = b

	return c
}
