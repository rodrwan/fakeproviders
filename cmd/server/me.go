package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/rodrwan/fakeproviders/repository/jwt"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func checkSession(ctx *Context, r *http.Request) (*jwt.Session, error) {
	sessSvc := jwt.SessionService{
		SecretKey: ctx.sessionSecretKey,
		MaxAge:    time.Duration(ctx.sessionMaxAge) * time.Second,
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		return nil, errors.New("invalid token")
	}

	sess, err := sessSvc.Session(r.Context(), &jwt.SessionCredentials{
		AuthToken: strings.Replace(token, "Bearer ", "", -1),
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func me(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	sess, err := checkSession(ctx, r)
	if err != nil {
		return &response{
			Data:   err.Error(),
			Status: http.StatusUnauthorized,
		}, nil
	}

	var userCard *card
	for _, card := range ctx.cards {
		if card.ID == sess.UserID {
			userCard = card
		}
	}

	return &response{
		Data:   userCard,
		Status: http.StatusOK,
	}, nil
}

func verify(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	sess, err := checkSession(ctx, r)
	if err != nil {
		return &response{
			Data:   err.Error(),
			Status: http.StatusUnauthorized,
		}, nil
	}

	authKey := StringWithCharset(12, charset)
	ctx.AuthKeys[sess.UserID] = authKey

	go func(id string) {
		time.Sleep(30 * time.Second)
		fmt.Println("delete auth key")
		delete(ctx.AuthKeys, id)
	}(sess.UserID)

	return &response{
		Data:   authKey,
		Status: http.StatusCreated,
	}, nil
}

func getCard(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	sess, err := checkSession(ctx, r)
	if err != nil {
		return &response{
			Data:   err.Error(),
			Status: http.StatusUnauthorized,
		}, nil
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var payload struct {
		VerificationToken string `json:"verification_token"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	verificationToken := payload.VerificationToken
	keyValue := ctx.AuthKeys[sess.UserID]
	if keyValue == "" {
		return &response{
			Data:   "a valid verification key must be provided",
			Status: http.StatusBadRequest,
		}, nil
	}

	if keyValue != verificationToken {
		return &response{
			Data:   "invalid verification token",
			Status: http.StatusBadRequest,
		}, nil
	}

	var userCard *card

	for _, card := range ctx.cards {
		if card.ID == sess.UserID {
			userCard = card
		}
	}
	cvv := fmt.Sprintf("%s%s%s", string(userCard.RealPAN[3]), string(userCard.RealPAN[7]), string(userCard.RealPAN[11]))

	return &response{
		Data: struct {
			NameOnCard string `json:"name_on_card"`
			PAN        string `json:"card_number"`
			ExpDate    string `json:"expiry_date"`
			CVV        string `json:"cvv"`
		}{
			NameOnCard: userCard.NameOnCard,
			PAN:        userCard.RealPAN,
			ExpDate:    userCard.RealExpDate,
			CVV:        cvv,
		},
		Status: http.StatusOK,
	}, nil
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
