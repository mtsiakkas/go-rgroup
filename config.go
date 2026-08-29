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
// All global configuration is set by calling methods on Config, and must be set
// before calling Config.Lock. Envelope response options are grouped under
// Config.Envelope.
var Config configSetter
var config = defaultConfig

func checkLock() {
	if config.locked {
		panic("[rgroup] config mutation after Config.Lock")
	}
}

// Lock locks the global config. Any further mutation of it panics.
// The config must be locked before any HandlerGroup or HandlerMux is created.
func (c configSetter) Lock() {
	config.locked = true
}

// Unlock unlocks the global config, so it can be mutated again.
func (c configSetter) Unlock() {
	config.locked = false
}

var defaultConfig = globalConfig{
	logOptions:      false,
	envelope:        envelopeOptions{},
	logger:          defaultLogger,
	prewriter:       nil,
	forwardErrorLog: false,
	recoverPanics:   true,
}

// Enable enables envelope responses, wrapping the handler response in an
// Envelope. A response whose Data is a []byte is written as-is and is never
// enveloped.
// Default: disabled.
func (e envelopeSetter) Enable() {
	checkLock()
	config.envelope.enabled = true
}

// Disable disables envelope responses.
// Default: disabled.
func (e envelopeSetter) Disable() {
	checkLock()
	config.envelope.enabled = false
}

// SetForwardLogMessage sets whether the log message of a response or error is
// sent to the client, as the message field of the Envelope status.
// Default: false
func (e envelopeSetter) SetForwardLogMessage(b bool) {
	checkLock()
	config.envelope.forwardLogMessage = b
}

// SetForwardHTTPStatus sets whether the handler's http status code is sent to
// the client. When false, enveloped responses are written with 200 OK and the
// status is reported only in the Envelope.
// Default: false
func (e envelopeSetter) SetForwardHTTPStatus(b bool) {
	checkLock()
	config.envelope.forwardHTTPStatus = b
}

// SetGlobalLogger sets the logger function called once per request, after the
// response has been written. Passing nil disables logging.
// A single route can override it with HandlerGroup.SetLogger.
func (c configSetter) SetGlobalLogger(p func(*LoggerData)) {
	checkLock()
	if p == nil {
		p = func(l *LoggerData) {}
	}

	config.logger = p
}

// SetLogOptionsRequests sets whether the logger function is called for OPTIONS
// requests.
// Default: false
func (c configSetter) SetLogOptionsRequests(b bool) {
	checkLock()
	config.logOptions = b
}

// SetPrewriter sets the global prewriter function, which is given the response
// before it is written to the client and returns the response to write.
// It can be used to further process every response in one place.
// It is not called when the Handler returns an error.
func (c configSetter) SetPrewriter(f func(*http.Request, *HandlerResponse) *HandlerResponse) {
	checkLock()
	config.prewriter = f
}

// SetRecoverPanics sets whether panics raised by handlers are recovered and
// reported to the client as 500 Internal Server Error.
// The recovered value and its stack trace are wrapped in a PanicError, carried
// by the HandlerError passed to the logger.
// When disabled, panics propagate to net/http, which aborts the connection
// without calling the logger.
// Default: true
func (c configSetter) SetRecoverPanics(b bool) {
	checkLock()
	config.recoverPanics = b
}

// SetForwardErrorLog sets whether the error's log message is appended to the
// client response of a HandlerError.
// It is only respected if envelope responses are not enabled, and has no effect
// on an error with no client response, since nothing is written for one.
// Default: false
func (c configSetter) SetForwardErrorLog(b bool) {
	checkLock()
	config.forwardErrorLog = b
}

func ensureLocked() {
	if !config.locked {
		panic("[rgroup] config not locked")
	}
}
