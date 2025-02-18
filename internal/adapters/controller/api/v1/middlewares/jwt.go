package middlewares

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"prod/cmd/app"
	"prod/internal/adapters/database/postgres"
	"prod/internal/adapters/database/redis"
	"prod/internal/domain/dto"
	"prod/internal/domain/entity"
	"prod/internal/domain/service"
	"prod/internal/domain/utils/auth"
)

type UserService interface {
	GetByID(ctx context.Context, uuid string) (*entity.User, error)
}

type BusinessService interface {
	GetByID(ctx context.Context, uuid string) (*entity.Business, error)
}

type TokenService interface {
	VerifyAuthId(ctx context.Context, userID, authID string) (bool, error)
}

type MiddlewareHandler struct {
	userService     UserService
	businessService BusinessService
	tokenService    TokenService
}

// NewMiddlewareHandler is a function that returns a new instance of MiddlewareHandler.
func NewMiddlewareHandler(app *app.App) *MiddlewareHandler {
	userStorage := postgres.NewUserStorage(app.DB)
	userService := service.NewUserService(userStorage)
	businessStorage := postgres.NewBusinessStorage(app.DB)
	businessService := service.NewBusinessService(businessStorage)

	tokenStorage := redis.NewTokenStorage(app.Redis)
	tokenService := service.NewTokenService(tokenStorage)

	return &MiddlewareHandler{
		userService:     userService,
		businessService: businessService,
		tokenService:    tokenService,
	}
}

// IsAuthenticated is a function that checks whether the user has sufficient rights to access the endpoint
/*
* tokenType string - the type of token that is required to access the endpoint
 */
func (h MiddlewareHandler) IsAuthenticated() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		user, authID, fetchErr := auth.GetUserFromJWT(authHeader, "access", c.Context(), h.userService.GetByID)

		business, businessAuthID, businessFetchErr := auth.GetBusinessFromJWT(authHeader, "access", c.Context(), h.businessService.GetByID)

		userVerify, userVerifyErr := h.tokenService.VerifyAuthId(c.Context(), user.ID, authID)

		businessVerify, businessVerifyErr := h.tokenService.VerifyAuthId(c.Context(), business.ID, businessAuthID)

		if (userVerifyErr != nil && businessVerifyErr != nil) || (!userVerify && !businessVerify) {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Время действия токена истекло.",
			})
		}

		if (fetchErr != nil && businessFetchErr != nil) || (user == nil && business == nil) {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Пользователь не авторизован.",
			})
		}

		if fetchErr == nil && user != nil {
			c.Locals("user", user)
		}
		if businessFetchErr == nil && business != nil {
			c.Locals("business", business)
		}

		return c.Next()
	}
}
