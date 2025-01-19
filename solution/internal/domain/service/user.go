package service

import (
	"context"
	"github.com/biter777/countries"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"strings"
)

type userStorage interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id string) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type userService struct {
	storage userStorage
}

func NewUserService(storage userStorage) *userService {
	return &userService{storage: storage}
}

func (s *userService) Create(ctx context.Context, registerReq dto.UserRegister) (*entity.User, error) {

	user := entity.User{
		Email:     registerReq.Email,
		Password:  nil,
		Name:      registerReq.Name,
		Surname:   registerReq.Surname,
		AvatarURL: registerReq.AvatarURL,
		Age:       registerReq.Other.Age,
		Country:   countries.ByName(strings.ToUpper(registerReq.Other.Country)),
	}
	user.SetPassword(registerReq.Password)
	return s.storage.Create(ctx, user)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.storage.GetByEmail(ctx, email)
}

func (s *userService) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return s.storage.GetByID(ctx, id)
}

func (s *userService) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return s.storage.Update(ctx, user)
}
