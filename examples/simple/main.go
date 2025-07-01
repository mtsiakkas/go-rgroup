package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtsiakkas/go-rgroup"
)

func main() {

	// Define handler goup
	g := rgroup.NewWithHandlers(rgroup.HandlerMap{
		http.MethodGet:  handleGet,
		http.MethodPost: handlePost,
	})

	// Generate http.HandlerFunc from HandlerGroup
	h := g.Make()

	// Create new http.ServeMux
	r := http.NewServeMux()

	// Add generated http.HandlerFunc to r
	r.HandleFunc("/", h)

	// Start http server
	fmt.Println("listening on localhost:3000")
	if err := http.ListenAndServe("localhost:3000", r); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}

}

// rgroup.Handler for GET
func handleGet(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	res := rgroup.Response("hello from GET handler").
		WithMessage("GET request - said hello").
		WithHTTPStatus(http.StatusAccepted)

	return res, nil
}

// rgroup.Handler for POST
// http.StatusNotImplemented error
func handlePost(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	err := rgroup.Error(http.StatusNotImplemented).
		WithResponse("POST method not implemented").
		WithMessage("POST request - not implemented")

	return nil, err
}
