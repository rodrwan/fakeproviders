package main

import (
	"fmt"
	"log"
	"net/http"
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

	c := newCard(&create.user)
	ctx.cards = append(ctx.cards, c)

	return &response{
		Status: http.StatusCreated,
		Data:   c,
	}, nil

}
