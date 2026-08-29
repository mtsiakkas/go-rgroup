// Basic rgroup.HandlerGroup example, listening on localhost:3000.
//
// Valid paths:
//
//	GET     /g1  202 "hello from GET 1 handler"
//	POST    /g1  501 "POST 1 method not implemented"
//	GET     /g2  202 "hello from GET 2 handler"
//	POST    /g2  501 "POST 2 method not implemented"
//	GET     /g3  200 "hello from the standalone handler"
//
// /g1 and /g2 are groups, so they answer OPTIONS with "Allow: GET,OPTIONS,POST"
// and any other method with 405 Method Not Allowed.
//
// /g3 is a single rgroup.Handler converted with Handler.ToHandlerFunc. It has no
// group around it to dispatch on method, so it answers every method alike.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtsiakkas/go-rgroup"
)

func main() {
	// The global config must be locked before any group is created.
	rgroup.Config.Lock()

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
	r := http.NewServeMux()

	// A HandlerGroup is an http.Handler, so it can be registered directly
	r.Handle("/g1", g1)
	r.Handle("/g2", g2)

	// A single rgroup.Handler can be served without a group by converting it
	// to an http.HandlerFunc
	r.HandleFunc("/g3", rgroup.Handler(handleGet3).ToHandlerFunc())

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

// Standalone rgroup.Handler, served without a group
func handleGet3(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
	res := rgroup.Response("hello from the standalone handler").
		WithMessage("GET 3 request - said hello")

	return res, nil
}
