package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/rodrwan/fakeproviders/logger"

	"github.com/ulule/limiter/drivers/middleware/stdlib"

	apierror "github.com/rodrwan/fakeproviders/api-error"
	corsLib "github.com/rs/cors"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"

	cardServer "github.com/rodrwan/fakeproviders/cmd/cards/server"
	userServer "github.com/rodrwan/fakeproviders/cmd/users/server"
)

const (
	minCreateProcessTime = 2
	maxCreateProcessTime = 10
	minLoadProcessTime   = 2
	maxLoadProcessTime   = 10
)

var (
	port  = flag.String("port", "8080", "Service port")
	token = flag.String("token", "fasdfadfa9fj987afsdf", "Token for authenticated endpointds")
)

func main() {
	flag.Parse()

	cc := &Context{}
	rate := limiter.Rate{
		Period: 10 * time.Second,
		Limit:  2,
	}
	store := memory.NewStore()

	// middlewares
	fakeLogger := logger.NewLogger("fakeprovider")
	rateLimitMid := stdlib.NewMiddleware(
		limiter.New(store, rate),
		stdlib.WithForwardHeader(true),
		stdlib.WithLimitReachedHandler(func(w http.ResponseWriter, r *http.Request) {
			apierror.NewError("Limit exceeded", http.StatusTooManyRequests).Write(w)
		}),
	)
	auth := NewAuthMiddleware(*token)

	r := NewRouter()
	r.GET("/", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, getAllCardsHandler})))
	r.POST("/cards", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, create})))
	r.POST("/load", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, loadHandler})))
	r.PATCH("/cards/:id/info", fakeLogger.Handle(auth.Handle(ContextHandler{cc, patch})))

	r.POST("/login", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, createSession})))
	r.GET("/api/me", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, me})))
	r.POST("/api/me/verify", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, verify})))
	r.POST("/api/me/card", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, getCard})))

	log.Printf("server running on %s", fmt.Sprintf(":%s", *port))

	cors := corsLib.New(corsLib.Options{
		AllowedOrigins:     []string{"*"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "Credentials"},
		AllowedMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		Debug:              true,
	})

	mux := http.NewServeMux()
	mux.Handle("/", cors.Handler(r))
	// running card service
	go cardServer.Run("8081")
	go userServer.Run("8082")

	panic(http.ListenAndServe(fmt.Sprintf(":%s", *port), mux))
}

func unmarshalJSON(r io.ReadCloser, v interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
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
