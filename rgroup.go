// A zero-dependency handler groupping framework for net/http.
package rgroup

import (
	"encoding/json"
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

type writeError struct {
	error
}

func writeErr(w http.ResponseWriter, err *HandlerError) (int, error) {
	if err == nil {
		return 0, nil
	}

	if Config.envelopeResponse != nil {
		env := err.ToEnvelope()

		return write(w, env)
	}

	w.WriteHeader(err.HTTPStatus)

	if err.Response != "" {
		n, e := w.Write([]byte(err.Response))
		if e != nil {
			return n, writeError{e}
		}
	}

	return 0, nil
}

func writeRes(w http.ResponseWriter, res *HandlerResponse) (int, error) {
	if res == nil {
		return 0, nil
	}

	if Config.envelopeResponse != nil && reflect.TypeFor[[]byte]() != reflect.TypeOf(res.Data) {
		env := res.ToEnvelope()

		if Config.envelopeResponse.forwardHTTPStatus && (res.HTTPStatus != http.StatusOK) {
			w.WriteHeader(res.HTTPStatus)
		}

		n, err := write(w, env)
		if err != nil {
			return 0, err
		}

		return n, nil
	}

	if res.HTTPStatus != http.StatusOK {
		w.WriteHeader(res.HTTPStatus)
	}

	n, err := write(w, res.Data)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func write(w http.ResponseWriter, d any) (int, error) {
	if d == nil {
		return 0, nil
	}

	var n int
	var err error

	switch reflect.TypeOf(d) {
	case reflect.TypeFor[string]():
		n, err = w.Write([]byte(d.(string))) //nolint
	case reflect.TypeFor[[]byte]():
		n, err = w.Write(d.([]byte)) //nolint
	default:
		dj, derr := json.Marshal(d)
		if derr != nil {
			return 0, writeError{derr}
		}

		n, err = w.Write(dj)
	}

	if err != nil {
		return 0, writeError{err}
	}

	return n, nil
}

func toPtr[T any](t T) *T {
	return &t
}
