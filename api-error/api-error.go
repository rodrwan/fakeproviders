package apierror

import (
	"encoding/json"
	"net/http"
)

// Response ...
type Response struct {
	Status int      `json:"-"`
	Error  *Message `json:"error,omitempty"`
}

// Message ...
type Message struct {
	Message string `json:"message"`
}

// Write writes a ApplicationResposne to the given response writer encoded as JSON.
func (er *Response) Write(w http.ResponseWriter) error {
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
func NewError(msg string, status int) *Response {
	return &Response{
		Status: status,
		Error: &Message{
			Message: msg,
		},
	}
}
