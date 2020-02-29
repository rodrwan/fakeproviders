package users

import (
	"errors"

	"github.com/rodrwan/fakeproviders/domain"
)

type IService interface {
	Create(u *domain.User) error
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
}

type Service struct {
	BaseURL   string
	AuthToken string
}

func (svc *Service) Create(u *domain.User) error {
	return errors.New("not implemented")
}

func (svc *Service) GetByID(id string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (svc *Service) GetByEmail(email string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}
