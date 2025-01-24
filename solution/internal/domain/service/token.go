package service

import (
	"context"
	"github.com/spf13/viper"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/utils/auth"
	"time"
)

type TokenStorage interface {
	SetAuthID(ctx context.Context, userID, authID string) error
	CheckAuthID(ctx context.Context, userID, authID string) (bool, error)
}

// tokenService is a struct that contains a pointer to a gorm.DB instance to interact with token repository.
type tokenService struct {
	storage TokenStorage
}

func NewTokenService(storage TokenStorage) *tokenService {
	return &tokenService{storage: storage}
}

// GenerateToken is a method to generate a new token.
func (s *tokenService) GenerateToken(ctx context.Context, userID string, expires time.Time, tokenType string) (*entity.Token, error) {
	token, id, err := auth.GenerateToken(userID, expires, tokenType)

	if err != nil {
		return nil, err
	}

	err = s.storage.SetAuthID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return &entity.Token{
		Token:   token,
		UserID:  userID,
		Type:    tokenType,
		Expires: expires,
	}, nil
}

// GenerateAuthTokens is a method to generate access and refresh tokens.
func (s *tokenService) GenerateAuthTokens(c context.Context, userID string) (*dto.AuthTokens, error) {
	authToken, err := s.GenerateToken(
		c,
		userID,
		time.Now().UTC().Add(time.Minute*time.Duration(viper.GetInt("service.backend.jwt.access-token-expiration"))),
		auth.TokenTypeAccess,
	)
	if err != nil {
		return nil, err
	}

	//refreshToken, err := s.GenerateToken(
	//	c,
	//	userID,
	//	time.Now().UTC().Add(time.Minute*time.Duration(viper.GetInt("service.backend.jwt.refresh-token-expiration"))),
	//	auth.TokenTypeRefresh,
	//)
	//if err != nil {
	//	return nil, err
	//}

	return &dto.AuthTokens{
		Access: dto.Token{
			Token:   authToken.Token,
			Expires: authToken.Expires,
		},
		//Refresh: dto.Token{
		//	Token:   refreshToken.Token,
		//	Expires: refreshToken.Expires,
		//},
	}, nil
}

func (s *tokenService) VerifyAuthId(ctx context.Context, userID, authID string) (bool, error) {
	return s.storage.CheckAuthID(ctx, userID, authID)
}
