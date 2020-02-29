package server

import (
	"fmt"
	corsLib "github.com/rs/cors"
	"log"
	"net/http"

	"github.com/rodrwan/fakeproviders/logger"
)

func Run(port string) {
	cc := &Context{}
	// middlewares
	fakeLogger := logger.NewLogger("fakeprovider")

	r := NewRouter()

	r.POST("/users", fakeLogger.Handle(ContextHandler{cc, createUser}))
	r.GET("/users/:id", fakeLogger.Handle(ContextHandler{cc, getUser}))

	log.Printf("server running on %s", fmt.Sprintf(":%s", port))

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
	panic(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}
