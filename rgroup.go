// A zero-dependency handler groupping framework for net/http.
package rgroup

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

func defaultLogger(r *LoggerData) {
	if r.Error != nil {
		log.Printf("\033[31m%s\033[0m", r)
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

	switch reflect.TypeOf(d) {
	case reflect.TypeFor[string]():
		n, err = w.Write([]byte(d.(string)))
	case reflect.TypeFor[[]byte]():
		n, err = w.Write(d.([]byte))
	default:
		dj, derr := json.Marshal(d)
		if derr != nil {
			err = derr
			break
		}
		n, err = w.Write(dj)
	}

	if err != nil {
		fmt.Printf("\033[31m[rgroup] failed to write to client: %s\n\033[0m", err)
	}

	return n
}

func toPtr[T any](t T) *T {
	return &t
}
