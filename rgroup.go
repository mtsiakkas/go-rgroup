// A zero-dependency handler groupping framework for net/http.
package rgroup

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
)

const (
	red   = "\033[31m"
	reset = "\033[0m"
)

var errorLogger *log.Logger = log.New(os.Stderr, red, log.LstdFlags)

func defaultLogger(r *LoggerData) {
	if r.Error != nil {
		errorLogger.Printf("%s%s", r, reset)
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

	if _, ok := res.Data.([]byte); !ok && Config.envelopeResponse != nil {
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
		errorLogger.Printf("[rgroup] failed to write to client: %s\n%s", err, reset)
	}

	return n
}

func logAndWrite(w http.ResponseWriter, l *LoggerData, logger func(*LoggerData)) {

	defer func() {
		if (l.Request.Method != http.MethodOptions || Config.logOptions) && logger != nil {
			l.Duration()
			logger(l)
		}
	}()

	if l.err != nil {
		me := new(HandlerError)
		if !errors.As(l.err, &me) {
			me.HTTPStatus = http.StatusInternalServerError
			_ = me.Wrap(l.err)
		}

		n := writeErr(w, me)

		l.Error = me
		l.ResponseSize = n

		return
	}

	if Config.prewriter != nil {
		l.Response = Config.prewriter(&l.Request, l.Response)
	}

	n := writeRes(w, l.Response)

	l.ResponseSize = n
}

func toPtr[T any](t T) *T {
	return &t
}
