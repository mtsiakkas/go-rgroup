package rgroup

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
)

func print(ctx context.Context, r *RequestData) {
	printFunc := log.Printf
	if r.IsError {
		printFunc = func(s string, args ...any) { log.Printf("\033[31m"+s+"\033[0m", args...) }
	}

	dur := float32(r.Duration)
	i := 0
	units := []string{"ns", "us", "ms", "s"}
	for dur > 1000 && i < 3 {
		dur /= 1000
		i++
	}

	if r.Message != "" {
		printFunc("%s %d %s [%3.1f%s]\n%s", r.Request.Method, r.Status, r.Path, dur, units[i], r.Message)
	} else {
		printFunc("%s %d %s [%3.1f%s]", r.Request.Method, r.Status, r.Path, dur, units[i])
	}
}

func write(w http.ResponseWriter, d any) (int, error) {
	n := 0
	var err error
	if d != nil {
		switch reflect.TypeOf(d) {
		case reflect.TypeFor[string]():
			n, err = w.Write([]byte(d.(string)))
		case reflect.TypeFor[[]byte]():
			n, err = w.Write(d.([]byte))
		default:
			dj, jerr := json.Marshal(d)
			err = jerr
			if jerr == nil {
				n, err = w.Write(dj)
			}
		}
	}
	if err != nil {
		return 0, err
	}
	return n, nil
}
