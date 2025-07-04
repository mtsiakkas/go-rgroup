package rgroup

import (
	"context"
	"fmt"
	"sync"
)

// GlobalConfig defines all global configuration options
type GlobalConfig struct {
	overwriteMethodBehaviour         OverwriteMethodBehaviour
	overwriteOptionsHandlerBehaviour OverwriteOptionsHandlerBehaviour
	postprocessOptions               bool
	envelopeResponse                 bool
	forwardHTTPStatus                bool
	forwardLogMessage                bool
	requestPostProcessor             func(context.Context, *RequestData)
}

var mtx = sync.RWMutex{}

var defaultConfig = GlobalConfig{
	postprocessOptions:               true,
	overwriteMethodBehaviour:         OverwriteMethodPanic,
	overwriteOptionsHandlerBehaviour: OverwriteOptionsHandlerPanic,
	envelopeResponse:                 false,
	forwardHTTPStatus:                false,
	forwardLogMessage:                false,
	requestPostProcessor:             nil,
}

// Config is a global instance of GlobalConfig and holds the global configuration for the package
var Config = defaultConfig

// Reset the global config to the default values
func (c *GlobalConfig) Reset() *GlobalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	*c = defaultConfig

	return c
}

// OverwriteMethodBehaviour defines what should happen if the Handler for a method is reassigned
type OverwriteMethodBehaviour int

/*
OverwriteMethodPanic - panic (default).
OverwriteMethodError - return error.
OverwriteMethodAllow - replace existing Handler.
OverwriteMethodIgnore - ignore new Handler keeping old.
*/
const (
	OverwriteMethodPanic OverwriteMethodBehaviour = iota
	OverwriteMethodIgnore
	OverwriteMethodAllow
	OverwriteMethodError
)

var duplicateMethodOpts = map[OverwriteMethodBehaviour]string{
	OverwriteMethodPanic:  "panic",
	OverwriteMethodIgnore: "ignore",
	OverwriteMethodAllow:  "allow",
	OverwriteMethodError:  "error",
}

// Validate - ensure d is a valid OverwriteMethodBehaviour option
func (d OverwriteMethodBehaviour) Validate() bool {
	_, ok := duplicateMethodOpts[d]

	return ok
}

// Implement Stringer interface
func (d OverwriteMethodBehaviour) String() string {
	return duplicateMethodOpts[d]
}

// OverwriteMethodUknownOptionError - simple error struct
// returned by SetOverwriteMethodBehaviour when passed option is invalid
type OverwriteMethodUknownOptionError struct {
	option OverwriteMethodBehaviour
}

func (e OverwriteMethodUknownOptionError) Error() string {
	return fmt.Sprintf("unknown OverwriteMethodBehaviour option %d", e.option)
}

// SetOverwriteMethodBehaviour - defines duplicate method behaviour
// returns OverwriteMethodUknownOptionError error if invalid option is passed.
func (c *GlobalConfig) SetOverwriteMethodBehaviour(o OverwriteMethodBehaviour) error {
	mtx.Lock()
	defer mtx.Unlock()

	if !o.Validate() {
		return OverwriteMethodUknownOptionError{option: o}
	}

	c.overwriteMethodBehaviour = o

	return nil
}

// GetOverwriteMethod - return current duplicate method setting
func (c *GlobalConfig) GetOverwriteMethod() OverwriteMethodBehaviour {
	mtx.RLock()
	defer mtx.RUnlock()

	return c.overwriteMethodBehaviour
}

// OverwriteOptionsHandlerBehaviour defines what should happen if the OPTIONS handler is manually set
type OverwriteOptionsHandlerBehaviour int

/*
OverwriteOptionsHandlerPanic - panic (default).
OverwriteOptionsHandlerIgnore - ignore new Handler keeping old.
OverwriteOptionsHandlerOverwrite - replace default handler.
*/
const (
	OverwriteOptionsHandlerPanic OverwriteOptionsHandlerBehaviour = iota // default
	OverwriteOptionsHandlerIgnore
	OverwriteOptionsHandlerOverwrite
)

var optsOpts = map[OverwriteOptionsHandlerBehaviour]string{
	OverwriteOptionsHandlerPanic:     "panic",
	OverwriteOptionsHandlerIgnore:    "ignore",
	OverwriteOptionsHandlerOverwrite: "overwrite",
}

// Implement Stringer interface
func (o OverwriteOptionsHandlerBehaviour) String() string {
	return optsOpts[o]
}

// Validate - ensure d is a valid OptionsHandlerBehaviour option
func (o OverwriteOptionsHandlerBehaviour) Validate() bool {
	_, ok := optsOpts[o]

	return ok
}

// OptionsHandlerUknownOptionError - error struct returned by
// SetOverwriteOptionsHandlerBehaviour when passed option is invalid
type OptionsHandlerUknownOptionError struct {
	option OverwriteOptionsHandlerBehaviour
}

func (e OptionsHandlerUknownOptionError) Error() string {
	return fmt.Sprintf("unknown OverwriteOptionsHandlerBehaviour option %d", e.option)
}

// SetOverwriteOptionsHandlerBehaviour - set options method overwrite setting
func (c *GlobalConfig) SetOverwriteOptionsHandlerBehaviour(o OverwriteOptionsHandlerBehaviour) error {
	mtx.Lock()
	defer mtx.Unlock()

	if !o.Validate() {
		return OptionsHandlerUknownOptionError{option: o}
	}

	c.overwriteOptionsHandlerBehaviour = o

	return nil
}

// GetOverwriteOptionsHandlerBehaviour - return the current options method overwrite behaviour
func (c *GlobalConfig) GetOverwriteOptionsHandlerBehaviour() OverwriteOptionsHandlerBehaviour {
	mtx.RLock()
	defer mtx.RUnlock()

	return c.overwriteOptionsHandlerBehaviour
}

// SetGlobalPostprocessor - set global request post processor
func (c *GlobalConfig) SetGlobalPostprocessor(p func(context.Context, *RequestData)) *GlobalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.requestPostProcessor = p

	return c
}

// GetGlobalPostprocessor - get global request post processor
func (c *GlobalConfig) GetGlobalPostprocessor() func(context.Context, *RequestData) {
	mtx.RLock()
	defer mtx.RUnlock()

	return c.requestPostProcessor
}

// SetPostprocessOptions - self explanaroty
func (c *GlobalConfig) SetPostprocessOptions(b bool) *GlobalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.postprocessOptions = b

	return c
}

// SetForwardLogMessage - self explanatory
func (c *GlobalConfig) SetForwardLogMessage(b bool) *GlobalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.forwardLogMessage = b

	return c
}

// SetForwardHTTPStatus - self explanatory
func (c *GlobalConfig) SetForwardHTTPStatus(b bool) *GlobalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.forwardHTTPStatus = b

	return c
}

// SetEnvelopeResponse - self explanatory
func (c *GlobalConfig) SetEnvelopeResponse(b bool) *GlobalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.envelopeResponse = b

	return c
}
