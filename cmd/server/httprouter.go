// Package router a wrapper over httprouter that is a trie based high performance
// HTTP request router.
//
// A trivial example is:
//
//  package main
//
//  import (
//  	"fmt"
//  	"log"
//  	"net/http"
//
//  	"github.com/Finciero/sigiriya/router"
//  )
//
//  func Index(w http.ResponseWriter, r *http.Request) {
//  	fmt.Fprint(w, "Welcome!\n")
//  }
//
//  func Hello(w http.ResponseWriter, r *http.Request) {
//  	ctx := r.Context()
//  	fmt.Fprintf(w, "hello, %s!\n", ctx.Value("name"))
//  }
//
//  func main() {
//  	router := router.New()
//  	router.GET("/", http.HandlerFunc(Index))
//  	router.GET("/hello/:name", http.HandlerFunc(Hello))
//  	log.Fatal(http.ListenAndServe(":8000", router))
//  }
//
// As you can see just changes the signature of the method provided by httprouter, and
// injects URI parameters in the request.Context.
package main

import (
	"net/http"

	"context"

	"github.com/julienschmidt/httprouter"
)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	httprouter.Router
}

// NewRouter returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func NewRouter() *Router {
	return &Router{
		Router: httprouter.Router{
			RedirectTrailingSlash:  true,
			RedirectFixedPath:      true,
			HandleMethodNotAllowed: true,
			HandleOPTIONS:          true,
		},
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handler http.Handler) {
	r.Handle("GET", path, handler)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handler http.Handler) {
	r.Handle("HEAD", path, handler)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handler http.Handler) {
	r.Handle("OPTIONS", path, handler)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handler http.Handler) {
	r.Handle("POST", path, handler)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handler http.Handler) {
	r.Handle("PUT", path, handler)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handler http.Handler) {
	r.Handle("PATCH", path, handler)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handler http.Handler) {
	r.Handle("DELETE", path, handler)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handler http.Handler) {
	r.Router.Handle(method, path, r.wrapHandler(handler))
}

func (r *Router) wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := r.Context()
		for _, p := range params {
			ctx = context.WithValue(ctx, p.Key, p.Value)
		}
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}
