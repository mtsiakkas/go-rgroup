package rgroup

import (
	"net/http"
	"sync"
)

type globalConfig struct {
	logOptions       bool
	envelopeResponse *envelopeOptions
	logger           func(*LoggerData)
	prewriter        func(*http.Request, *HandlerResponse) *HandlerResponse
}

type envelopeOptions struct {
	forwardHTTPStatus bool
	forwardLogMessage bool
}

var mtx = sync.RWMutex{}

var defaultConfig = globalConfig{
	logOptions:       true,
	envelopeResponse: nil,
	logger:           nil,
	prewriter:        nil,
}

// Config holds the global configuration for the package.
// All global configurations are set by calling methods on Config.
var Config globalConfig

func init() {
	Config = defaultConfig
}

// Reset the global config to the default values.
func (c *globalConfig) Reset() *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	*c = defaultConfig

	return c
}

// Set the global logger function.
func (c *globalConfig) SetGlobalLogger(p func(*LoggerData)) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.logger = p

	return c
}

// Call logger function on OPTIONS requests.
// Default: true
func (c *globalConfig) SetLogOptionsRequests(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.logOptions = b

	return c
}

// Forward the log message to the client.
// Calling this method automatically enables envelope responses.
// Default: false
func (c *globalConfig) SetForwardLogMessage(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	if c.envelopeResponse == nil {
		c.envelopeResponse = new(envelopeOptions)
	}

	c.envelopeResponse.forwardLogMessage = b

	return c
}

// Forward http status code to client.
// Calling this method automatically enables envelope responses.
// Default: false
func (c *globalConfig) SetForwardHTTPStatus(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	if c.envelopeResponse == nil {
		c.envelopeResponse = new(envelopeOptions)
	}

	c.envelopeResponse.forwardHTTPStatus = b

	return c
}

// Enable envelope responses.
// Default: false
func (c *globalConfig) SetEnvelopeResponse(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	if b {
		c.envelopeResponse = new(envelopeOptions)
	} else {
		c.envelopeResponse = nil
	}

	return c
}

// Set global prewriter function.
// This can be used to further process the response before writing to the client.
func (c *globalConfig) SetPrewriter(f func(*http.Request, *HandlerResponse) *HandlerResponse) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.prewriter = f

	return c
}
