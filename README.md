# rgroup
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mtsiakkas/go-rgroup?logo=go)
![GitHub Tag](https://img.shields.io/github/v/tag/mtsiakkas/go-rgroup)
![GitHub branch check runs](https://img.shields.io/github/check-runs/mtsiakkas/go-rgroup/main)
![GitHub License](https://img.shields.io/github/license/mtsiakkas/go-rgroup)

A zero-dependency handler groupping framework for net/http.

# Overview

rgroup is a framework to simplify both the structuring and implementation of APIs using the standard library net/http package. 

## Features
- Simple tuple return from handlers
- Per-route middleware
- Customizable request logger
- Builtin options handler
- Envelope responses
- User defined prewriter function

# Usage

```go
router := http.NewServeMux()

group := rgroup.New()

group.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
   return rgroup.Response("Hello World!").WithHTTPStatus(http.StatusAccepted), nil
})

group.Post( func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
   return nil, rgroup.Error(http.StatusNotImplemented).WithMessage("TODO")
})

router.Handle("/", group)
```

The route definition can be inlined using `rgroup.NewWithHandlers(...)`
```go
router := http.NewServeMux().Handle("/", rgroup.NewWithHandlers(rgroup.HandlerMap{
    http.MethodGet: func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
       return rgroup.Response("Hello World!"), nil
    },
    http.MethodPost:  func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
       return rgroup.Response("Hello World!"), nil
    },
}))
```

# Configuration
Configuration is set via `rgroup.Config`.

## Global logger
rgroup comes with a builtin request logger. This can be globally overwritten with 
```go

func logger(r *RequestData) {
    log.Printf("NEW REQUEST: %s",r)
}

rgroup.Config.SetGlobalLogger(logger)
```

## Envelope responses
rgroup can be configured to envelope responses by calling `rgroup.Config.SetEnvelopeResponse(true)` responding to the client with a fixed structure json object
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

By default enveloped responses always return a `200 OK` code to the client. This can be changed with `rgroup.Config.SetForwardHTTPStatus(true)` to forward the  status code to the client.

## Log options requests
By default `OPTIONS` requests are not logged. This behaviour can be changed with `rgroup.Config.SetLogOptionsRequests(true)`.
