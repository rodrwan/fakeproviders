package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type createRequestData struct {
	user
}

func create(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	var create createRequestData

	defer r.Body.Close()
	if err := unmarshalJSON(r.Body, &create); err != nil {
		log.Println(fmt.Errorf("Error: %v", err))
		return nil, err
	}

	for _, c := range ctx.cards {
		if c.User.Email == create.Email {
			return nil, errors.New("user already have a card")
		}
	}

	c := newCard(&create.user)
	if err := randomError(); err != nil {
		return nil, err
	}
	ctx.cards = append(ctx.cards, c)

	return &response{
		Status: http.StatusCreated,
		Data:   c,
	}, nil

}

func randomError() error {
	rand.Seed(time.Now().UTC().UnixNano())
	prob := rand.Float64()

	if prob < 0.3 {
		log.Println("Something funny (:")
		return errors.New("Something went wrong")
	}

	return nil
}
