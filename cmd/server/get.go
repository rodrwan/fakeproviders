package main

import (
	"net/http"
)

func getAllCardsHandler(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	return &response{
		Status: http.StatusOK,
		Data:   ctx.cards,
	}, nil

}
