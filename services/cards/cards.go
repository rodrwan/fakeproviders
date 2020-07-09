package cards

import (
	"errors"

	"github.com/rodrwan/fakeproviders/domain"
)

type IService interface {
	Create(c *domain.Card) error
	GetByID(id string) (*domain.Card, error)
	GetByEmail(email string) (*domain.Card, error)
}

type Service struct {
	BaseURL   string
	AuthToken string
}

func (svc *Service) Create(u *domain.Card) error {
	return errors.New("not implemented")
}

func (svc *Service) GetByID(id string) (*domain.Card, error) {
	return nil, errors.New("not implemented")
}

func (svc *Service) GetByEmail(email string) (*domain.Card, error) {
	return nil, errors.New("not implemented")
}
