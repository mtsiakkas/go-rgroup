package utils_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/mtsiakkas/go-rgroup/internal/utils"
	"github.com/mtsiakkas/go-rgroup/pkg/request"
)

func TestPrint(t *testing.T) {
	r := request.RequestData{
		Id:           1,
		Path:         "/test",
		Params:       nil,
		Ts:           100,
		Method:       http.MethodGet,
		Duration:     200,
		Message:      "",
		Status:       http.StatusAccepted,
		IsError:      false,
		ResponseSize: 100,
	}

	res1 := captureOutput(func() { utils.Print(context.Background(), &r) })
	if !strings.HasSuffix(res1, "GET 202 /test [200.0ns]\n") {
		t.Logf("unexpected print output: %s", res1)
		t.Fail()
	}

	r.Message = "test message"
	r.Duration = 2300

	res2 := captureOutput(func() { utils.Print(context.Background(), &r) })
	if !strings.HasSuffix(res2, "GET 202 /test [2.3us]\ntest message\n") {
		t.Logf("unexpected print output: %s", res2)
		t.Fail()
	}

	r.IsError = true
	res3 := captureOutput(func() { utils.Print(context.Background(), &r) })
	if !strings.HasSuffix(res3, "\033[31mGET 202 /test [2.3us]\ntest message\033[0m\n") {
		t.Logf("unexpected print output: %s", res3)
		t.Fail()
	}
}

func TestWrite(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		rr := httptest.NewRecorder()

		n, err := utils.Write(rr, "test")
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if n != 4 {
			t.Logf("unexpected response length: %d", n)
			t.Fail()
		}

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if string(b) != "test" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}
	})

	t.Run("[]byte", func(t *testing.T) {
		rr := httptest.NewRecorder()

		n, err := utils.Write(rr, []byte("test"))
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if n != 4 {
			t.Logf("unexpected response length: %d", n)
			t.Fail()
		}

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if string(b) != "test" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}
	})

	t.Run("struct", func(t *testing.T) {
		rr := httptest.NewRecorder()

		n, err := utils.Write(rr, struct {
			Data string `json:"data"`
		}{Data: "test"})
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if n != 15 {
			t.Logf("unexpected response length: %d", n)
			t.Fail()
		}

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if string(b) != "{\"data\":\"test\"}" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}
	})
}

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
		log.SetOutput(os.Stderr)
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
