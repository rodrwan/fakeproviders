package main

import "net/http"

func personalData(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	return &response{
		Status: http.StatusOK,
		Data:   nil,
	}, nil
}
