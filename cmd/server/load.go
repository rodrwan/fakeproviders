package main

import (
	"log"
	"net/http"
	"time"
)

type loadRequestData struct {
	ReferenceID string `json:"reference_id"`
	Amount      int64  `json:"amount"`
}

func loadHandler(ctx *Context, w http.ResponseWriter, r *http.Request) (*response, error) {
	var load loadRequestData
	defer r.Body.Close()
	if err := unmarshalJSON(r.Body, &load); err != nil {
		return nil, err
	}

	processTime := randomProcessTime(minLoadProcessTime, maxLoadProcessTime) * time.Second
	log.Printf("Waiting for %.2fs", processTime.Seconds())
	time.Sleep(processTime)

	var selectedCard *card
	for _, card := range ctx.cards {
		if card.ReferenceID == load.ReferenceID {
			card.Balance += load.Amount
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
