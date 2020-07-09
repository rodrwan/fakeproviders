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

	defer r.Body.Close()
	if err := unmarshalJSON(r.Body, &patch); err != nil {
		log.Println(fmt.Errorf("Error: %v", err))
		return nil, err
	}

	var selectedCard *card
	for _, card := range ctx.cards {
		if card.ID == id {
			card.PAN = patch.CardNumber
			card.ExpDate = patch.ExpDate
			card.CVV = patch.CVV
			card.ReferenceID = patch.ReferenceID
			card.UpdatedAt = time.Now()

			selectedCard = card
			break
		}
	}

	if selectedCard == nil {
		return &response{
			Status: http.StatusNotFound,
		}, nil
	}

	return &response{
		Status: http.StatusOK,
		Data:   selectedCard,
	}, nil

}
