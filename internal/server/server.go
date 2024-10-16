package server

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/leomfn/rueckenwind/internal/middleware"
)

type server struct {
	address string
	mux     *http.ServeMux
}

func NewServer(port int64) *server {
	// TODO: maybe move port validation here

	return &server{
		address: fmt.Sprintf(":%d", port),
		mux:     http.NewServeMux(),
	}
}

func (s *server) AddRouter(router *router) {
	s.mux.Handle(router.path, router.mux)
}

func (s *server) Start() {
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

// TODO: Maybe use as server method, which automatically binds a router to a
// server
func NewRouter(path string) *router {
	return &router{
		path: path,
		mux:  http.NewServeMux(),
	}
}

// Register a new handler, with optional middleware(s). The handler is wrapped
// by the middlewares in reverse order they are provided. The allowed method
// must be supplied, if all methods are allowed, then 'ALL' must be passed.
func (r *router) Handle(method string, path string, handler http.Handler, middlewares ...middleware.Middleware) {
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
		handler = m.MiddlewareFunc(handler)
	}

	r.mux.Handle(methodPath, handler)
}
