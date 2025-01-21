package b2b

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"solution/cmd/app"
	"solution/internal/adapters/controller/api/validator"
	"solution/internal/adapters/database/postgres"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"time"
)

type BusinessService interface {
	Create(ctx context.Context, registerReq dto.BusinessRegister) (*entity.Business, error)
	GetByID(ctx context.Context, uuid string) (*entity.Business, error)
	Update(ctx context.Context, business *entity.Business) (*entity.Business, error)
	GetByEmail(ctx context.Context, email string) (*entity.Business, error)
}

type TokenService interface {
	GenerateAuthTokens(c context.Context, businessID string) (*dto.AuthTokens, error)
	GenerateToken(ctx context.Context, businessID string, expires time.Time, tokenType string) (*entity.Token, error)
}

type BusinessHandler struct {
	businessService BusinessService
	tokenService    TokenService
	validator       *validator.Validator
}

func NewBusinessHandler(app *app.App) *BusinessHandler {
	businessStorage := postgres.NewBusinessStorage(app.DB)
	tokenStorage := postgres.NewTokenStorage(app.DB)

	return &BusinessHandler{
		businessService: service.NewBusinessService(businessStorage),
		tokenService:    service.NewTokenService(tokenStorage),
		validator:       app.Validator,
	}
}

func (h BusinessHandler) register(c fiber.Ctx) error {
	var businessDTO dto.BusinessRegister

	if err := c.Bind().Body(&businessDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(businessDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	business, errCreate := h.businessService.Create(c.Context(), businessDTO)
	if errCreate != nil {
		return c.Status(fiber.StatusConflict).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Такой email уже зарегистрирован.",
		})
	}

	tokens, tokensErr := h.tokenService.GenerateAuthTokens(c.Context(), business.ID)
	if tokensErr != nil || tokens == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "failed to generate auth tokens",
		})
	}

	response := dto.BusinessRegisterResponse{
		BusinessID: business.ID,
		Token:      tokens.Access.Token,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h BusinessHandler) login(c fiber.Ctx) error {
	var businessDTO dto.BusinessLogin

	if err := c.Bind().Body(&businessDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(businessDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	business, errFetch := h.businessService.GetByEmail(c.Context(), businessDTO.Email)
	if errFetch != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Неверный email или пароль.",
		})
	}

	passErr := business.ComparePassword(businessDTO.Password)
	if passErr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Неверный email или пароль.",
		})
	}

	tokens, tokensErr := h.tokenService.GenerateAuthTokens(c.Context(), business.ID)
	if tokensErr != nil || tokens == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "failed to generate auth tokens",
		})
	}

	response := dto.BusinessLoginResponse{
		Token: tokens.Access.Token,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h BusinessHandler) Setup(router fiber.Router) {
	businessAuthGroup := router.Group("/business/auth")
	businessAuthGroup.Post("/sign-up", h.register)
	businessAuthGroup.Post("/sign-in", h.login)
}
