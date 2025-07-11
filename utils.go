// Package rgroup - A handler groupper for go
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

func write(w http.ResponseWriter, d any) (int, error) {
	if d == nil {
		return 0, nil
	}

	switch reflect.TypeOf(d) {
	case reflect.TypeFor[string]():
		return w.Write([]byte(d.(string))) //nolint
	case reflect.TypeFor[[]byte]():
		return w.Write(d.([]byte)) //nolint
	default:
		dj, err := json.Marshal(d)
		if err != nil {
			return 0, writeError{err}
		}

		n, err := w.Write(dj)
		if err != nil {
			return 0, writeError{err}
		}

		return n, nil
	}
}

func toPtr[T any](t T) *T {
	return &t
}
