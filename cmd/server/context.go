package main

import "net/http"

type Context struct {
	cards []*card
}

type ContextHandler struct {
	ctx *Context
	H   func(*Context, http.ResponseWriter, *http.Request) (*response, error)
}

// Our ServeHTTP method is mostly the same, and also has the ability to
// access our *appContext's fields (templates, loggers, etc.) as well.
func (ah ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Updated to pass ah.appContext as a parameter to our handler type.
	resp, err := ah.H(ah.ctx, w, r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	switch resp.Status {
	case http.StatusNotFound:
		http.NotFound(w, r)
	case http.StatusInternalServerError:
		http.Error(w, http.StatusText(resp.Status), resp.Status)
	}

	resp.Write(w)
}
