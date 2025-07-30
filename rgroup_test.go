package rgroup

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

func captureErrorLog(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stderr := os.Stderr
	defer func() {
		os.Stderr = stderr
		errorLogger.SetOutput(os.Stderr)
	}()
	os.Stderr = writer
	errorLogger.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	_ = writer.Close()
	return <-out
}

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
		log.SetOutput(os.Stdout)
	}()
	os.Stdout = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	_ = writer.Close()
	return <-out
}

func TestDefaultLogger(t *testing.T) {
	l := fromRequest(*httptest.NewRequest(http.MethodGet, "/", nil))

	r := captureOutput(func() { defaultLogger(l) })
	if r == "" {
		t.Logf("no logger output")
		t.Fail()
	}
	if !strings.Contains(r, "GET 200 /") {
		t.Logf("unexpected logger output: %s", r)
		t.Fail()
	}

	l.Error = Error(http.StatusInternalServerError).WithMessage("test error")
	r = captureErrorLog(func() { defaultLogger(l) })
	if r == "" {
		t.Logf("no logger output")
		t.Fail()
	}
	if !strings.Contains(r, "GET 500 /") {
		t.Logf("unexpected logger output: %s", r)
		t.Fail()
	}
}

func TestWriteErr(t *testing.T) {
	rr := httptest.NewRecorder()
	n := writeErr(rr, nil)
	if n != 0 {
		t.Logf("unexpected message length: %d", n)
		t.Fail()
	}
	if rr.Body.String() != "" {
		t.Logf("unexpected error message: %s", rr.Body.String())
		t.Fail()
	}

	rr = httptest.NewRecorder()
	err := Error(http.StatusNotAcceptable).WithMessage("test error")

	writeErr(rr, err)
	if rr.Code != http.StatusNotAcceptable {
		t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
		t.Fail()
	}

	err.WithResponse("test response")
	rr = httptest.NewRecorder()

	writeErr(rr, err)
	if rr.Body.String() != "test response" {
		t.Logf("unexpected error message: %s", rr.Body.String())
		t.Fail()
	}

	Config.SetEnvelopeResponse(true)
	rr = httptest.NewRecorder()

	writeErr(rr, err)
	if rr.Body.String() != "{\"status\":{\"http_status\":406,\"error\":\"test response\"}}" {
		t.Logf("unexpected error message: %s", rr.Body.String())
		t.Fail()
	}

	Config.SetForwardLogMessage(true)
	rr = httptest.NewRecorder()

	writeErr(rr, err)
	if rr.Body.String() != "{\"status\":{\"http_status\":406,\"message\":\"test error\",\"error\":\"test response\"}}" {
		t.Logf("unexpected error message: %s", rr.Body.String())
		t.Fail()
	}

	Config.Reset()
}

func TestWriteRes(t *testing.T) {
	rr := httptest.NewRecorder()

	res := Response("test data").
		WithMessage("test message").
		WithHTTPStatus(http.StatusAccepted).
		WithHeader("X-Test-1", "test1").
		WithHeader("X-Test-2", "test2")

	writeRes(rr, res)

	if rr.Code != http.StatusAccepted {
		t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
		t.Fail()
	}
	if rr.Body.String() != "test data" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}
	if rr.Header().Get("X-Test-1") != "test1" || rr.Header().Get("X-Test-2") != "test2" {
		t.Log("unexpected headers:")
		t.Logf("X-Test-1: %s", rr.Header().Get("X-Test-1"))
		t.Logf("X-Test-2: %s", rr.Header().Get("X-Test-2"))
		t.Fail()
	}

	Config.SetEnvelopeResponse(true)
	rr = httptest.NewRecorder()
	writeRes(rr, res)
	if rr.Body.String() != "{\"data\":\"test data\",\"status\":{\"http_status\":202}}" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}

	Config.SetForwardHTTPStatus(true)
	rr = httptest.NewRecorder()

	writeRes(rr, res)
	if rr.Code != http.StatusAccepted {
		t.Logf("unexpected status code: %d (%s)", rr.Code, http.StatusText(rr.Code))
		t.Fail()
	}
	if rr.Body.String() != "{\"data\":\"test data\",\"status\":{\"http_status\":202}}" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}

	Config.Reset()
}

func TestWrite(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		rr := httptest.NewRecorder()
		n := write(rr, "test string")
		if n == 0 {
			t.Log("no bytes written")
			t.Fail()
		}

		if rr.Body.String() != "test string" {
			t.Logf("unexpected response: %s", rr.Body.String())
			t.Fail()
		}
	})

	t.Run("bytes", func(t *testing.T) {
		rr := httptest.NewRecorder()
		n := write(rr, []byte("test string"))
		if n == 0 {
			t.Log("no bytes written")
			t.Fail()
		}

		if rr.Body.String() != "test string" {
			t.Logf("unexpected response: %s", rr.Body.String())
			t.Fail()
		}
	})

	t.Run("struct", func(t *testing.T) {
		s := struct {
			Data string `json:"data"`
			Len  int    `json:"len"`
		}{Data: "test string", Len: 123}

		rr := httptest.NewRecorder()
		n := write(rr, s)
		if n == 0 {
			t.Log("no bytes written")
			t.Fail()
		}

		if rr.Body.String() != "{\"data\":\"test string\",\"len\":123}" {
			t.Logf("unexpected response: %s", rr.Body.String())
			t.Fail()
		}
	})

	t.Run("nil", func(t *testing.T) {
		rr := httptest.NewRecorder()
		write(rr, nil)

		if rr.Body.String() != "" {
			t.Logf("unexpected body: %s", rr.Body.String())
			t.Fail()
		}
	})

	t.Run("error", func(t *testing.T) {
		ew := ErrorWriter{}
		log := captureErrorLog(func() { write(ew, "test error") })
		if !strings.HasSuffix(log, "[rgroup] failed to write to client: test error\n\033[0m\n") {
			t.Logf("unexpected output: %s", log)
			t.Fail()
		}
	})
}

type ErrorWriter struct {
}

func (w ErrorWriter) Header() http.Header        { return nil }
func (w ErrorWriter) Write([]byte) (int, error)  { return 0, errors.New("test error") }
func (w ErrorWriter) WriteHeader(statusCode int) {}
