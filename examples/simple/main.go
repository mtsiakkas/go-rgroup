package main

import (
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
	h2 := g2.Make()

	// Create new http.ServeMux
	r := http.NewServeMux()

	// Add generated http.Handler/http.HandlerFunc to r
	r.Handle("/g1", g1)
	r.HandleFunc("/g2", h2)

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
	err := rgroup.Error(http.StatusNotImplemented).
		WithResponse("POST 2 method not implemented").
		WithMessage("POST 2 request - not implemented")

	return nil, err
}
