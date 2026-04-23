package main

import (
	"fmt"
	"net/http"

	"github.com/mtsiakkas/go-rgroup"
)

func main() {

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

	r3.HandleFunc("/g3/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello from g3"))
		w.WriteHeader(http.StatusBadRequest)
	})
	r2.Handle("/r3/", http.StripPrefix("/r3", r3))

	r.Handle("/r2/", r2.SetPrefix("/r2").AddMiddleware(middleware))

	// Start http server
	fmt.Println("listening on localhost:3000")
	http.ListenAndServe("localhost:3000", r)
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
