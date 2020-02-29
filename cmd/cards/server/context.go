package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	apierror "github.com/rodrwan/fakeproviders/api-error"
	"github.com/rodrwan/fakeproviders/services/cards"
)

// Context context holds shared data between services and handlers
type Context struct {
	// CardsServices is a service that store Card information in an agnostig database service.
	CardsService cards.IService
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
	case http.StatusBadRequest:
		data := resp.Data.(string)
		apierror.NewError(data, http.StatusBadRequest).Write(w)

		return
	case http.StatusUnauthorized:
		apierror.NewError("", http.StatusUnauthorized).Write(w)
		return
	case http.StatusNotFound:
		http.NotFound(w, r)
		return
	case http.StatusInternalServerError:
		apierror.NewError(err.Error(), http.StatusInternalServerError).Write(w)
		return
	}

	resp.Write(w)
}

type response struct {
	Status int         `json:"-"`
	Data   interface{} `json:"data,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
}

// Write writes a ApplicationResposne to the given response writer encoded as JSON.
func (r *response) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)
	_, err = w.Write(b)
	return err
}

type errorResponse struct {
	Status int           `json:"-"`
	Error  *errorMessage `json:"error,omitempty"`
}

type errorMessage struct {
	Message string `json:"message"`
}

// Write writes a ApplicationResposne to the given response writer encoded as JSON.
func (er *errorResponse) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(er)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(er.Status)
	_, err = w.Write(b)
	return err
}

// NewError ...NewError
func NewError(msg string, status int) *errorResponse {
	return &errorResponse{
		Status: status,
		Error: &errorMessage{
			Message: msg,
		},
	}
}

func unmarshalJSON(r io.ReadCloser, v interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
