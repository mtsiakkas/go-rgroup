package rgroup

import (
	"net/http"
)

type globalConfig struct {
	logOptions      bool
	envelope        envelopeOptions
	logger          func(*LoggerData)
	prewriter       func(*http.Request, *HandlerResponse) *HandlerResponse
	forwardErrorLog bool
	recoverPanics   bool
	locked          bool
}

type envelopeOptions struct {
	enabled           bool
	forwardHTTPStatus bool
	forwardLogMessage bool
}

type envelopeSetter struct{}
type configSetter struct{ Envelope envelopeSetter }

// Config holds the global configuration for the package.
// All global configurations are set by calling methods on Config.
var Config configSetter
var config = defaultConfig

func checkLock() {
	if config.locked {
		panic("[rgroup] config mutation after Config.Lock")
	}
}

// Lock global config
// Any further mutations panic
func (c configSetter) Lock() {
	config.locked = true
}

// Unlock global config
func (c configSetter) Unlock() {
	config.locked = false
}

func resetConfig() {
	config = defaultConfig
}

var defaultConfig = globalConfig{
	logOptions:      false,
	envelope:        envelopeOptions{},
	logger:          defaultLogger,
	prewriter:       nil,
	forwardErrorLog: false,
	recoverPanics:   true,
}

// Enable envelope response. Disabled by default
func (e envelopeSetter) Enable() {
	checkLock()
	config.envelope.enabled = true
}

// Disable envelope response. Disabled by default
func (e envelopeSetter) Disable() {
	checkLock()
	config.envelope.enabled = false
}

// Forward the log message to the client.
// Default: false
func (e envelopeSetter) SetForwardLogMessage(b bool) {
	checkLock()
	config.envelope.forwardLogMessage = b
}

// Forward http status code to client.
// Default: false
func (e envelopeSetter) SetForwardHTTPStatus(b bool) {
	checkLock()
	config.envelope.forwardHTTPStatus = b
}

// Set the global logger function.
func (c configSetter) SetGlobalLogger(p func(*LoggerData)) {
	checkLock()
	if p == nil {
		p = func(l *LoggerData) {}
	}

	config.logger = p
}

// Call logger function on OPTIONS requests.
// Default: false
func (c configSetter) SetLogOptionsRequests(b bool) {
	checkLock()
	config.logOptions = b
}

// Set global prewriter function.
// This can be used to further process the response before writing to the client.
func (c configSetter) SetPrewriter(f func(*http.Request, *HandlerResponse) *HandlerResponse) {
	checkLock()
	config.prewriter = f
}

// Recover panics raised by handlers and respond with 500 Internal Server Error.
// The recovered value and its stack trace are wrapped in a PanicError, carried
// by the HandlerError passed to the logger.
// When disabled, panics propagate to net/http, which aborts the connection
// without calling the logger.
// Default: true
func (c configSetter) SetRecoverPanics(b bool) {
	checkLock()
	config.recoverPanics = b
}

// Send error log message to client.
// This is only respected if envelope responses are not enabled.
func (c configSetter) SetForwardErrorLog(b bool) {
	checkLock()
	config.forwardErrorLog = b
}

func ensureLocked() {
	if !config.locked {
		panic("[rgroup] config not locked")
	}
}
