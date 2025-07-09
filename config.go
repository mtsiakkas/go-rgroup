package rgroup

import (
	"net/http"
	"sync"
)

// globalConfig defines all global configuration options
type globalConfig struct {
	logOptions       bool
	envelopeResponse *envelopeOptions
	logger           func(*RequestData)
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

// Config is a global instance of GlobalConfig and holds the global configuration for the package
var Config = defaultConfig

// Reset the global config to the default values
func (c *globalConfig) Reset() *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	*c = defaultConfig

	return c
}

// SetGlobalLogger - set global request post processor
func (c *globalConfig) SetGlobalLogger(p func(*RequestData)) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.logger = p

	return c
}

// GetGlobalLogger - get global request post processor
func (c *globalConfig) GetGlobalLogger() func(*RequestData) {
	mtx.RLock()
	defer mtx.RUnlock()

	return c.logger
}

// SetLogOptionsRequests - self explanaroty
func (c *globalConfig) SetLogOptionsRequests(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.logOptions = b

	return c
}

// SetForwardLogMessage - self explanatory
func (c *globalConfig) SetForwardLogMessage(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	if c.envelopeResponse == nil {
		c.envelopeResponse = new(envelopeOptions)
	}

	c.envelopeResponse.forwardLogMessage = b

	return c
}

// SetForwardHTTPStatus - self explanatory
func (c *globalConfig) SetForwardHTTPStatus(b bool) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	if c.envelopeResponse == nil {
		c.envelopeResponse = new(envelopeOptions)
	}

	c.envelopeResponse.forwardHTTPStatus = b

	return c
}

// SetEnvelopeResponse - self explanatory
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

// SetPrewriter - self explanatory
func (c *globalConfig) SetPrewriter(f func(*http.Request, *HandlerResponse) *HandlerResponse) *globalConfig {
	mtx.Lock()
	defer mtx.Unlock()

	c.prewriter = f

	return c
}
