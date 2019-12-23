package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rodrwan/fakeproviders/repository/jwt"
)

func createSession(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	if payload.Username != ctx.username || payload.Password != ctx.password {
		return &response{
			Status: http.StatusBadRequest,
			Data:   "invalid username or password",
		}, nil
	}

	// create jwt
	sess, err := jwt.NewSession(payload.Username, ctx.userUUID, r.Header.Get("Origin"))
	if err != nil {
		return nil, err
	}

	sessSvc := jwt.SessionService{
		SecretKey: ctx.sessionSecretKey,
		MaxAge:    time.Duration(ctx.sessionMaxAge) * time.Second,
	}

	creds, err := sessSvc.CreateSession(r.Context(), sess)
	if err != nil {
		return nil, err
	}

	return &response{
		Status: http.StatusCreated,
		Data:   creds.AuthToken,
	}, nil
}
