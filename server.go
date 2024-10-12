package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	port int64
)

// Load server environment variables
func init() {
	var err error

	portEnv, exists := os.LookupEnv("PORT")
	if !exists {
		log.Fatal("PORT environment variable not set")
	}

	// TODO: maybe 32 bit is enough, probably irrelevant
	port, err = strconv.ParseInt(portEnv, 10, 64)

	if err != nil {
		log.Fatal("PORT environment variable must be an integer")
	}
}

type server struct {
	address string
	mux     *http.ServeMux
	path    string
}

func newServer() *server {
	return &server{
		address: fmt.Sprintf(":%d", port),
		mux:     http.NewServeMux(),
	}
}

func (s *server) addRouter(router *router) {
	s.mux.Handle(router.path, router.mux)
}

func (s *server) start() {
	server := &http.Server{
		Addr:    s.address,
		Handler: s.mux,
		// TODO: maybe add ReadTimeout, WriteTimeout
	}

	log.Printf("Starting server on %v", s.address)
	log.Fatal(server.ListenAndServe())
}

type router struct {
	path string
	mux  *http.ServeMux
}

func newRouter(path string) *router {
	return &router{
		path: path,
		mux:  http.NewServeMux(),
	}
}

// Register a new handler, with optional middleware(s). The handler is wrapped
// by the middlewares in reverse order they are provided. The allowed method
// must be supplied, if all methods are allowed, then 'ALL' must be passed.
func (r *router) handle(method string, path string, handler http.Handler, middlewares ...middlewareFunc) {
	methodPath := strings.TrimSuffix(r.path, "/") + path

	switch method {
	case "ALL":
		break
	// define allowed methods
	case "GET", "POST":
		methodPath = fmt.Sprintf("%s %s", method, methodPath)
	default:
		log.Fatalf("Invalid method when registering handler: %s", method)
	}

	for _, m := range slices.Backward(middlewares) {
		handler = m(handler)
	}

	r.mux.Handle(methodPath, handler)
}
