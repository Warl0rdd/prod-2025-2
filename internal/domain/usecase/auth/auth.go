package auth

import (
	"context"
	"prod/internal/domain/dto"
	"prod/internal/domain/entity"
	"time"
)

type TokenService interface {
	GenerateToken(ctx context.Context, userID string, expires time.Time, tokenType string) (*entity.Token, error)
	DeleteToken(ctx context.Context, userID string, tokenType string) error
	GenerateAuthTokens(c context.Context, userID string) (*dto.AuthTokens, error)
}

type UserService interface {
	GetByEmail(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, registerReq dto.UserRegister) (*entity.User, error)
}
