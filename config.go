package rgroup

import (
	"net/http"
	"sync"
)

type globalConfig struct {
	logOptions      bool
	Envelope        envelopeOptions
	logger          func(*LoggerData)
	prewriter       func(*http.Request, *HandlerResponse) *HandlerResponse
	forwardErrorLog bool
	lockOnMake      bool
}

type envelopeOptions struct {
	enabled           bool
	forwardHTTPStatus bool
	forwardLogMessage bool
}

var mtx = sync.Mutex{}

var defaultConfig = globalConfig{
	logOptions:      true,
	Envelope:        envelopeOptions{},
	logger:          defaultLogger,
	prewriter:       nil,
	forwardErrorLog: false,
	lockOnMake:      true,
}

// Enable envelope response. Disabled by default
func (e *envelopeOptions) Enable() {
	mtx.Lock()
	defer mtx.Unlock()

	e.enabled = true
}

// Disable envelope response. Disabled by default
func (e *envelopeOptions) Disable() {
	mtx.Lock()
	defer mtx.Unlock()

	e.enabled = false
}

// Forward the log message to the client.
// Default: false
func (e *envelopeOptions) SetForwardLogMessage(b bool) {
	mtx.Lock()
	defer mtx.Unlock()

	e.forwardLogMessage = b
}

// Forward http status code to client.
// Default: false
func (e *envelopeOptions) SetForwardHTTPStatus(b bool) {
	mtx.Lock()
	defer mtx.Unlock()

	e.forwardHTTPStatus = b

}

// Config holds the global configuration for the package.
// All global configurations are set by calling methods on Config.
var Config globalConfig = defaultConfig

// Reset the global config to the default values.
func (c *globalConfig) Reset() {
	mtx.Lock()
	defer mtx.Unlock()

	*c = defaultConfig
}

// Set the global logger function.
func (c *globalConfig) SetGlobalLogger(p func(*LoggerData)) {
	mtx.Lock()
	defer mtx.Unlock()

	if p == nil {
		p = func(l *LoggerData) {}
	}

	c.logger = p
}

// Call logger function on OPTIONS requests.
// Default: true
func (c *globalConfig) SetLogOptionsRequests(b bool) {
	mtx.Lock()
	defer mtx.Unlock()

	c.logOptions = b
}

// Set global prewriter function.
// This can be used to further process the response before writing to the client.
func (c *globalConfig) SetPrewriter(f func(*http.Request, *HandlerResponse) *HandlerResponse) {
	mtx.Lock()
	defer mtx.Unlock()

	c.prewriter = f
}

var lockOnMakeOnce sync.Once

// Lock HandlerGroup after the first call to HandlerGroup.Make.
// Can only be called once.
// Default: true
func (c *globalConfig) LockOnMake(b bool) {
	mtx.Lock()
	defer mtx.Unlock()

	lockOnMakeOnce.Do(func() {
		c.lockOnMake = b
	})
}

// Send error log message to client.
// This is only respected if envelope responses are not enabled.
func (c *globalConfig) SetForwardErrorLog(b bool) {
	mtx.Lock()
	defer mtx.Unlock()

	c.forwardErrorLog = b
}
