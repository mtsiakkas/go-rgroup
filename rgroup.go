// A zero-dependency handler groupping framework for net/http.
package rgroup

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

	if Config.Envelope.enabled {
		env := err.ToEnvelope()

		if Config.Envelope.forwardHTTPStatus {
			w.WriteHeader(err.HTTPStatus)
		}

		return write(w, env)
	}

	w.WriteHeader(err.HTTPStatus)

	res := err.Response
	if errLog := err.Error(); Config.forwardErrorLog && errLog != "" {
		res = fmt.Sprintf("%s: %s", res, errLog)
	}

	if err.Response != "" {
		return write(w, res)
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

	if _, ok := res.Data.([]byte); !ok && Config.Envelope.enabled {
		env := res.ToEnvelope()

		if Config.Envelope.forwardHTTPStatus && (res.HTTPStatus != http.StatusOK) {
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
		if l.Request.Method != http.MethodOptions || Config.logOptions {
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

type rwriter struct {
	data    []byte
	status  int
	headers http.Header
}

func (r *rwriter) Header() http.Header {
	return r.headers
}

func (r *rwriter) Write(b []byte) (int, error) {
	r.data = b
	return len(b), nil
}

func (r *rwriter) WriteHeader(statusCode int) {
	r.status = statusCode
}

func fromHandler(h http.Handler) Handler {
	return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		ww := new(rwriter)
		ww.headers = http.Header{}
		if ww.status < 100 {
			ww.status = http.StatusOK
		}

		h.ServeHTTP(ww, req)

		if ww.status > 399 {
			return nil, Error(ww.status).WithResponse(string(ww.data))
		} else {
			res := Response(ww.data).WithHTTPStatus(ww.status)
			for k, v := range ww.Header() {
				res.WithHeader(k, strings.Join(v, ","))
			}
			return res, nil
		}
	}
}

func toPtr[T any](t T) *T {
	return &t
}
