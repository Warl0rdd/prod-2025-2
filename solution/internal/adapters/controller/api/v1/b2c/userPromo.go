package b2c

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"solution/cmd/app"
	"solution/internal/adapters/controller/api/validator"
	"solution/internal/adapters/database/postgres"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"strconv"
)

type PromoService interface {
	GetFeed(ctx context.Context, user *entity.User, dto dto.PromoFeedRequest) ([]dto.PromoForUser, int64, error)
	GetByIdUser(ctx context.Context, promoID string) (dto.PromoForUser, error)
}

type UserPromoHandler struct {
	PromoService PromoService
	validator    *validator.Validator
}

func NewUserPromoHandler(app *app.App) *UserPromoHandler {
	promoStorage := postgres.NewPromoStorage(app.DB)
	businessStorage := postgres.NewBusinessStorage(app.DB)

	return &UserPromoHandler{
		PromoService: service.NewPromoService(promoStorage, businessStorage),
		validator:    app.Validator,
	}
}

func (h UserPromoHandler) GetFeed(c fiber.Ctx) error {
	var requestDTO dto.PromoFeedRequest

	if err := c.Bind().Query(&requestDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	user := c.Locals("user").(*entity.User)

	promos, total, err := h.PromoService.GetFeed(c.Context(), user, requestDTO)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Status:  "error",
			Message: err.Error(),
		})
	}

	c.Append("X-Total-Count", strconv.FormatInt(total, 10))

	return c.Status(fiber.StatusOK).JSON(promos)
}

func (h UserPromoHandler) GetPromoByID(c fiber.Ctx) error {
	var requestDTO dto.PromoGetByID

	if err := c.Bind().URI(&requestDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(requestDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if user := c.Locals("user"); user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Пользователь не авторизован.",
		})
	}

	promo, err := h.PromoService.GetByIdUser(c.Context(), requestDTO.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(promo)
}

func (h UserPromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	userGroup := router.Group("/user")
	userGroup.Get("/feed", h.GetFeed, middleware)
	userGroup.Get("/promo/:id", h.GetPromoByID, middleware)
}
