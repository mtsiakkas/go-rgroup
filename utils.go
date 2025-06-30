package rgroup

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
)

func timeScale(t int64) (float32, string) {
	dur := float32(t)
	i := 0
	units := []string{"ns", "us", "ms", "s"}
	for dur > 1000 && i < 3 {
		dur /= 1000
		i++
	}

	return dur, units[i]
}

func defaultPrint(ctx context.Context, r *RequestData) {
	printFunc := log.Printf
	if r.IsError {
		printFunc = func(s string, args ...any) { log.Printf("\033[31m"+s+"\033[0m", args...) }
	}

	printFunc(r.String())
}

func write(w http.ResponseWriter, d any) (int, error) {
	if d == nil {
		return 0, nil
	}

	switch reflect.TypeOf(d) {
	case reflect.TypeFor[string]():
		//nolint
		return w.Write([]byte(d.(string)))
	case reflect.TypeFor[[]byte]():
		//nolint
		return w.Write(d.([]byte))
	default:
		dj, err := json.Marshal(d)
		if err != nil {
			return 0, err
		}

		return w.Write(dj)
	}
}
