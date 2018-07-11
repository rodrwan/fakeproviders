package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type patchRequestData struct {
	CardNumber  string `json:"card_number"`
	ExpDate     string `json:"exp_date"`
	CVV         string `json:"cvv"`
	ReferenceID string `json:"reference_id"`
}

func patch(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	id, ok := r.Context().Value("id").(string)
	if !ok {
		return nil, errors.New("missing id")
	}
	var patch patchRequestData

	fmt.Println(id)
	defer r.Body.Close()
	if err := unmarshalJSON(r.Body, &patch); err != nil {
		log.Println(fmt.Errorf("Error: %v", err))
		return nil, err
	}

	var cc *card

	for _, c := range ctx.cards {
		if c.ID == id {
			c.PAN = patch.CardNumber
			c.ExpDate = patch.ExpDate
			c.CVV = patch.CVV
			c.ReferenceID = patch.ReferenceID
			c.UpdatedAt = time.Now()

			cc = c
			break
		}
	}

	return &response{
		Status: http.StatusOK,
		Data:   cc,
	}, nil

}
