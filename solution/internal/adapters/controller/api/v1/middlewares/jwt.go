package middlewares

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"solution/cmd/app"
	"solution/internal/adapters/database/postgres"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"solution/internal/domain/utils/auth"
)

type UserService interface {
	GetByID(ctx context.Context, uuid string) (*entity.User, error)
}

type BusinessService interface {
	GetByID(ctx context.Context, uuid string) (*entity.Business, error)
}

type MiddlewareHandler struct {
	userService     UserService
	businessService BusinessService
}

// NewMiddlewareHandler is a function that returns a new instance of MiddlewareHandler.
func NewMiddlewareHandler(app *app.App) *MiddlewareHandler {
	userStorage := postgres.NewUserStorage(app.DB)
	userService := service.NewUserService(userStorage)
	businessStorage := postgres.NewBusinessStorage(app.DB)
	businessService := service.NewBusinessService(businessStorage)

	return &MiddlewareHandler{
		userService:     userService,
		businessService: businessService,
	}
}

// IsAuthenticated is a function that checks whether the user has sufficient rights to access the endpoint
/*
* tokenType string - the type of token that is required to access the endpoint
 */
func (h MiddlewareHandler) IsAuthenticated() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		user, fetchErr := auth.GetUserFromJWT(authHeader, "access", c.Context(), h.userService.GetByID)

		business, businessFetchErr := auth.GetBusinessFromJWT(authHeader, "access", c.Context(), h.businessService.GetByID)

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
