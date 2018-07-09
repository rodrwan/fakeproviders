package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const (
	minCreateProcessTime = 10
	maxCreateProcessTime = 300
	minLoadProcessTime   = 3
	maxLoadProcessTime   = 10
)

func main() {
	r := mux.NewRouter()

	// This is where the router is useful, it allows us to declare methods that
	// this path will be valid for
	r.HandleFunc("/create", createHandler).Methods("POST")
	r.HandleFunc("/load", loadHandler).Methods("POST")

	// We can then pass our router (after declaring all our routes) to this method
	// (where previously, we were leaving the secodn argument as nil)
	port := os.Getenv("PORT")
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

type createRequestData struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	EmailName string `json:"email_name,omitempty"`
}

func unmarshalJSON(r io.ReadCloser, v interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var create createRequestData

	defer r.Body.Close()
	if err := unmarshalJSON(r.Body, &create); err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	processTime := randomProcessTime(minCreateProcessTime, maxCreateProcessTime) * time.Second

	time.Sleep(processTime)

	rr := &response{
		Data: "Hello World!",
	}

	rr.Write(w)
}

type loadRequestData struct {
	ReferenceID string `json:"reference_id,omitempty"`
	Amount      int64  `json:"amount,omitempty"`
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	processTime := randomProcessTime(minLoadProcessTime, maxLoadProcessTime) * time.Second

	time.Sleep(processTime)
	fmt.Fprintf(w, "Hello World!")
}

func randomProcessTime(min, max int) time.Duration {
	rand.Seed(time.Now().UTC().UnixNano())
	return time.Duration(rand.Intn(max-min) + min)
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
