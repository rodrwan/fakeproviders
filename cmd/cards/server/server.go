package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rodrwan/fakeproviders/logger"
	corsLib "github.com/rs/cors"
)

func Run(port string) {

	// should add external service to store data.
	cc := &Context{}
	// middlewares
	fakeLogger := logger.NewLogger("fakeprovider cards")

	r := NewRouter()
	r.POST("/cards", fakeLogger.Handle(ContextHandler{cc, createCard}))
	r.POST("/load", fakeLogger.Handle(ContextHandler{cc, loadCard}))
	r.PATCH("/cards/:id", fakeLogger.Handle(ContextHandler{cc, getCard}))

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
	mux.Handle("/api", cors.Handler(r))
	panic(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}
