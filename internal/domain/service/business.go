package service

import (
	"context"
	"prod/internal/domain/dto"
	"prod/internal/domain/entity"
)

type businessStorage interface {
	Create(ctx context.Context, business entity.Business) (*entity.Business, error)
	GetByID(ctx context.Context, id string) (*entity.Business, error)
	GetAll(ctx context.Context, limit, offset int) ([]entity.Business, error)
	Update(ctx context.Context, business *entity.Business) (*entity.Business, error)
	Delete(ctx context.Context, id string) error
	GetByEmail(ctx context.Context, email string) (*entity.Business, error)
}

type businessService struct {
	storage businessStorage
}

func NewBusinessService(storage businessStorage) *businessService {
	return &businessService{storage: storage}
}

func (s *businessService) Create(ctx context.Context, registerReq dto.BusinessRegister) (*entity.Business, error) {

	business := entity.Business{
		Email: registerReq.Email,
		Name:  registerReq.Name,
	}
	business.SetPassword(registerReq.Password)
	return s.storage.Create(ctx, business)
}

func (s *businessService) GetByEmail(ctx context.Context, email string) (*entity.Business, error) {
	return s.storage.GetByEmail(ctx, email)
}

func (s *businessService) GetByID(ctx context.Context, id string) (*entity.Business, error) {
	return s.storage.GetByID(ctx, id)
}

func (s *businessService) Update(ctx context.Context, business *entity.Business) (*entity.Business, error) {
	return s.storage.Update(ctx, business)
}
