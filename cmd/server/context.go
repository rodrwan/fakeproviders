package main

import (
	"net/http"

	apierror "github.com/rodrwan/fakeprovider/api-error"
)

// Context context holds shared data between services and handlers
type Context struct {
	cards []*card
}

// ContextHandler join context with handler signature
type ContextHandler struct {
	ctx *Context
	H   func(*Context, http.ResponseWriter, *http.Request) (*response, error)
}

// Our ServeHTTP method is mostly the same, and also has the ability to
// access our *appContext's fields (templates, loggers, etc.) as well.
func (ah ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := ah.H(ah.ctx, w, r)
	if err != nil {
		apierror.NewError(err.Error(), http.StatusInternalServerError).Write(w)
		return
	}

	switch resp.Status {
	case http.StatusNotFound:
		http.NotFound(w, r)
		return
	case http.StatusInternalServerError:
		apierror.NewError(err.Error(), http.StatusInternalServerError).Write(w)
		return
	}

	resp.Write(w)
}
