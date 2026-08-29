# rgroup
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mtsiakkas/go-rgroup?logo=go)
![GitHub Tag](https://img.shields.io/github/v/tag/mtsiakkas/go-rgroup)
![GitHub branch check runs](https://img.shields.io/github/check-runs/mtsiakkas/go-rgroup/main)
![GitHub License](https://img.shields.io/github/license/mtsiakkas/go-rgroup)

A zero-dependency handler grouping framework for net/http.

# Overview

rgroup is a framework to simplify both the structuring and implementation of APIs using the standard library net/http package.

## Features
- Simple tuple return from handlers
- Per-route and inherited middleware
- Customizable request logger
- Builtin options handler
- Envelope responses
- Panic recovery
- User defined prewriter function

# Usage

Global configuration must be locked with `rgroup.Config.Lock()` before any group
or mux is created. Building a group while the config is unlocked panics with
`[rgroup] config not locked`.

```go
rgroup.Config.Lock()

router := http.NewServeMux()

group := rgroup.New()

group.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
    return rgroup.Response("Hello World!").WithHTTPStatus(http.StatusAccepted), nil
})

group.Handle(http.MethodPost, func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
    return nil, rgroup.Error(http.StatusNotImplemented).WithMessage("TODO")
})

router.Handle("/", group)
```

`HandlerGroup.Handle` returns the group, so calls can be chained. The method is
normalized to upper case, and registering the same method twice panics.

The route definition can be inlined using `rgroup.NewWithHandlers(...)`
```go
rgroup.Config.Lock()

router := http.NewServeMux()
router.Handle("/", rgroup.NewWithHandlers(rgroup.HandlerMap{
    http.MethodGet: func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
        return rgroup.Response("Hello World!"), nil
    },
    http.MethodPost: func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
        return rgroup.Response("Hello World!"), nil
    },
}))
```

A `HandlerGroup` implements `http.Handler`, so it can be registered on any
router. All setup - handlers, middleware and the local logger - must be
completed before the group serves its first request; building a group after it
has served panics with `[rgroup] build after serve`.

`OPTIONS` is handled automatically unless a handler is registered for it,
responding with an `Allow` header listing the methods handled by the group. Any
other unhandled method responds with `405 Method Not Allowed` and the same
`Allow` header.

## Middleware
`Middleware` is a `func(rgroup.Handler) rgroup.Handler`. Middleware added to a
group applies to every handler in that group.

```go
func middleware(h rgroup.Handler) rgroup.Handler {
    return func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
        res, err := h(w, req)
        if err != nil {
            return nil, err
        }
        return res.WithHeader("X-Example", "1"), nil
    }
}

group.AddMiddleware(middleware)
```

Adding a nil middleware panics.

## HandlerMux
`rgroup.NewServeMux` returns a `HandlerMux`, a wrapper around `http.ServeMux`
that propagates its middleware to the handlers registered on it. Nested
`HandlerGroup` and `HandlerMux` values inherit the parent's middleware; a plain
`http.Handler` is adapted so the middleware applies to it as well.

```go
rgroup.Config.Lock()

mux := rgroup.NewServeMux()
mux.Handle("/hello/", group)

sub := rgroup.NewServeMux()
sub.Handle("/world/", other)

mux.Handle("/sub/", sub.SetPrefix("/sub").AddMiddleware(middleware))

http.ListenAndServe("localhost:3000", mux)
```

`SetPrefix` strips the given prefix before dispatching, so nested muxes can be
registered under a path without repeating it in their own routes. Registering
the same path twice, a nil handler, or an empty path panics.

# Configuration
Configuration is set via `rgroup.Config`, and applies to the whole package.
All configuration must be set before calling `rgroup.Config.Lock()`; mutating it
afterwards panics with `[rgroup] config mutation after Config.Lock`. Use
`rgroup.Config.Unlock()` to reopen the config for changes.

## Global logger
rgroup comes with a builtin request logger. This can be globally overwritten with
```go
func logger(l *rgroup.LoggerData) {
    log.Printf("NEW REQUEST: %s", l)
}

rgroup.Config.SetGlobalLogger(logger)
```

Passing `nil` disables logging. A single group can override the global logger
with `group.SetLogger(logger)`.

## Envelope responses
rgroup can be configured to envelope responses by calling `rgroup.Config.Envelope.Enable()`, responding to the client with a fixed structure json object
```js
{
    data: ...,
    status: {
        http_status: number,
        message?: string,
        error?: string
    }
}
```

`data` is omitted when the handler returns no data, and `error` is set from the
handler error's client response, falling back to the status text. Responses
whose `Data` is a `[]byte` are written as-is and are never enveloped.

By default enveloped responses always return a `200 OK` code to the client. This can be changed with `rgroup.Config.Envelope.SetForwardHTTPStatus(true)` to forward the status code to the client.

The log message of a response or error is not sent to the client by default.
`rgroup.Config.Envelope.SetForwardLogMessage(true)` includes it as `status.message`.

## Forward error log
When envelope responses are disabled, `rgroup.Config.SetForwardErrorLog(true)`
appends the error's log message to the response body sent to the client.
Default: `false`.

## Log options requests
By default `OPTIONS` requests are not logged. This behaviour can be changed with `rgroup.Config.SetLogOptionsRequests(true)`.

## Panic recovery
Panics raised by handlers are recovered and reported to the client as
`500 Internal Server Error`. The recovered value and its stack trace are wrapped
in a `rgroup.PanicError`, carried by the `HandlerError` passed to the logger.
`http.ErrAbortHandler` is never swallowed. Recovery can be disabled with
`rgroup.Config.SetRecoverPanics(false)`, in which case panics propagate to
net/http, which aborts the connection without calling the logger.
Default: `true`.

## Prewriter
`rgroup.Config.SetPrewriter(f)` registers a
`func(*http.Request, *rgroup.HandlerResponse) *rgroup.HandlerResponse` that is
given the response before it is written to the client. It is not called for
handler errors.
