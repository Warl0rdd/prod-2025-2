package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"prod/internal/domain/common/errorz"
	"prod/internal/domain/entity"
	"strings"
	"time"
)

func VerifyToken(authHeader, secret, tokenType string) (string, string, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return "", "", errorz.AuthHeaderIsEmpty
	}

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	jwtType, ok := claims["type"].(string)
	if !ok || jwtType != tokenType {
		return "", "", errors.New("invalid token type")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", "", errors.New("invalid token sub")
	}

	if time.Unix(int64(claims["exp"].(float64)), 0).Before(time.Now()) {
		return "", "", errors.New("token expired")
	}

	authID, ok := claims["jti"].(string)
	if !ok {
		return "", "", errors.New("invalid token jti")
	}

	return userID, authID, nil
}

func GetUserFromJWT(jwt, tokenType string, context context.Context, getUser func(context.Context, string) (*entity.User, error)) (*entity.User, string, error) {
	id, authID, errVerify := VerifyToken(jwt, viper.GetString("service.backend.jwt.secret"), tokenType)
	if errVerify != nil {
		return &entity.User{}, "", errVerify
	}

	user, errGetUser := getUser(context, id)
	if errGetUser != nil {
		return &entity.User{}, "", errGetUser
	}

	return user, authID, nil
}

func GetBusinessFromJWT(jwt, tokenType string, context context.Context, getBusiness func(context.Context, string) (*entity.Business, error)) (*entity.Business, string, error) {
	id, authID, errVerify := VerifyToken(jwt, viper.GetString("service.backend.jwt.secret"), tokenType)
	if errVerify != nil {
		return &entity.Business{}, "", errVerify
	}

	business, errGetUser := getBusiness(context, id)
	if errGetUser != nil {
		return &entity.Business{}, "", errGetUser
	}

	return business, authID, nil
}

func GenerateToken(userID string, expires time.Time, tokenType string) (string, string, error) {
	jti := uuid.New().String()
	claims := jwt.MapClaims{
		"sub":  userID,
		"jti":  jti,
		"iat":  time.Now().Unix(),
		"exp":  expires.Unix(),
		"type": tokenType,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(viper.GetString("service.backend.jwt.secret")))
	if err != nil {
		return "", "", err
	}

	return signed, jti, nil
}
