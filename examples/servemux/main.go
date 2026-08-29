// Nested rgroup.HandlerMux example, listening on localhost:3000.
//
// Valid paths, all GET:
//
//	/g1/        200 "hello from g1"        group on the root mux, no middleware
//	/r2/g2/     200 "mid: hello from g2"   nested group, inherits r2's middleware
//	/r2/r3/g3/  400 "hello from g3"        nested http.ServeMux, error passed
//	                                       through the middleware untouched
//
// The trailing slash matters: these are http.ServeMux subtree patterns, so
// requests without it are redirected.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtsiakkas/go-rgroup"
)

func main() {
	// The global config must be locked before any group or mux is created.
	rgroup.Config.Lock()

	// Create new rgroup.ServeMux
	r := rgroup.NewServeMux()
	// Create rgroup sub router
	r2 := rgroup.NewServeMux()
	// Create http sub router
	r3 := http.NewServeMux()

	// Define handler groups
	g1 := rgroup.NewWithHandlers(rgroup.HandlerMap{
		http.MethodGet: func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("hello from g1"), nil
		}})
	r.Handle("/g1/", g1)

	g2 := rgroup.NewWithHandlers(rgroup.HandlerMap{
		http.MethodGet: func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("hello from g2"), nil
		}})
	r2.Handle("/g2/", g2)

	// A plain http.Handler is adapted, so the mux middleware applies to it too.
	// This one fails, and the middleware passes the error through untouched.
	r3.HandleFunc("/g3/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("hello from g3"))
	})
	r2.Handle("/r3/", http.StripPrefix("/r3", r3))

	// Middleware added to r2 is inherited by everything registered on it,
	// so it wraps g2 and the adapted r3 handler
	r.Handle("/r2/", r2.SetPrefix("/r2").AddMiddleware(middleware))

	// Start http server
	fmt.Println("listening on localhost:3000")
	if err := http.ListenAndServe("localhost:3000", r); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

func middleware(h rgroup.Handler) rgroup.Handler {
	return func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
		res, err := h(w, req)
		if err != nil {
			return nil, err
		}
		switch d := res.Data.(type) {
		case string:
			return rgroup.Response("mid: " + d).WithHTTPStatus(res.HTTPStatus), nil
		case []byte:
			return rgroup.Response("mid: " + string(d)).WithHTTPStatus(res.HTTPStatus), nil
		}
		return res, err
	}
}
