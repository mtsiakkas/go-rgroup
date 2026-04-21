package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mtsiakkas/go-rgroup"
)

func main() {

	// Define handler groups
	g1 := rgroup.NewWithHandlers(rgroup.HandlerMap{
		http.MethodGet:  handleGet1,
		http.MethodPost: handlePost1,
	})

	g2 := rgroup.NewWithHandlers(rgroup.HandlerMap{
		http.MethodGet:  handleGet2,
		http.MethodPost: handlePost2,
	})

	// Create new http.ServeMux
	r := rgroup.NewServeMux()

	// Create sub router
	r2 := rgroup.NewServeMux()
	r2.Handle("/g2/", g2)

	// new http.Handler (http.ServeMux)
	r3 := http.NewServeMux()
	r3.HandleFunc("/g3-1/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("X-Rgroup", "TEST")
		w.Write([]byte("header test"))
		w.WriteHeader(http.StatusOK)
	})
	r3.HandleFunc("/g3-2/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("error test"))
		w.WriteHeader(http.StatusBadRequest)
	})

	r.Handle("/g1/", g1)
	r.Handle("/r2/", r2.SetPrefix("/r2"))
	r.Handle("/r3/", http.StripPrefix("/r3", r3))
	r.AddMiddleware(middleware)

	// Start http server
	fmt.Println("listening on localhost:3000")
	if err := http.ListenAndServe("localhost:3000", r); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

// rgroup.Handler for GET method on g1
func handleGet1(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	res := rgroup.Response("hello from GET 1 handler").
		WithMessage("GET 1 request - said hello").
		WithHTTPStatus(http.StatusAccepted)

	return res, nil
}

// rgroup.Handler for POST method on g1
// http.StatusNotImplemented error
func handlePost1(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	err := rgroup.Error(http.StatusNotImplemented).
		WithResponse("POST 1 method not implemented").
		WithMessage("POST 1 request - not implemented")

	return nil, err
}

// rgroup.Handler for GET method on g2
func handleGet2(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	res := rgroup.Response("hello from GET 2 handler").
		WithMessage("GET 2 request - said hello").
		WithHTTPStatus(http.StatusAccepted)

	return res, nil
}

// rgroup.Handler for POST method on g2
// http.StatusNotImplemented error
func handlePost2(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	ctx := req.Context()
	message := ctx.Value("mid-ctx")
	return rgroup.Response(message), nil
}

func middleware(h rgroup.Handler) rgroup.Handler {
	return func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
		ctx := context.WithValue(req.Context(), "mid-ctx", "hello from context middleware")
		fmt.Println("middleware run")
		return h(w, req.WithContext(ctx))
	}
}
