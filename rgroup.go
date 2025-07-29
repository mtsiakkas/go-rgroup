// A zero-dependency handler groupping framework for net/http.
package rgroup

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

const (
	red   = "\033[31m"
	reset = "\033[0m"
)

func defaultLogger(r *LoggerData) {
	if r.Error != nil {
		log.Printf("%s%s%s", red, r, reset)
	} else {
		log.Println(r)
	}
}

func writeErr(w http.ResponseWriter, err *HandlerError) int {
	if err == nil {
		return 0
	}

	if Config.envelopeResponse != nil {
		env := err.ToEnvelope()
		return write(w, env)
	}

	w.WriteHeader(err.HTTPStatus)

	if err.Response != "" {
		return write(w, []byte(err.Response))
	}

	return 0
}

func writeRes(w http.ResponseWriter, res *HandlerResponse) int {
	if res == nil {
		return 0
	}

	if len(res.Headers) > 0 {
		for h, v := range res.Headers {
			w.Header().Add(h, v)
		}
	}

	if Config.envelopeResponse != nil && reflect.TypeFor[[]byte]() != reflect.TypeOf(res.Data) {
		env := res.ToEnvelope()

		if Config.envelopeResponse.forwardHTTPStatus && (res.HTTPStatus != http.StatusOK) {
			w.WriteHeader(res.HTTPStatus)
		}

		return write(w, env)
	}

	if res.HTTPStatus != http.StatusOK {
		w.WriteHeader(res.HTTPStatus)
	}

	return write(w, res.Data)
}

func write(w http.ResponseWriter, d any) int {
	if d == nil {
		return 0
	}

	var n int
	var err error

	switch d := d.(type) {
	case string:
		n, err = w.Write([]byte(d))
	case []byte:
		n, err = w.Write(d)
	default:
		dj, derr := json.Marshal(d)
		if derr != nil {
			err = derr
			break
		}
		n, err = w.Write(dj)
	}

	if err != nil {
		fmt.Printf("%s[rgroup] failed to write to client: %s\n%s", red, err, reset)
	}

	return n
}

func toPtr[T any](t T) *T {
	return &t
}
