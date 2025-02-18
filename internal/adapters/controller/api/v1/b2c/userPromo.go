package b2c

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v3"
	"prod/cmd/app"
	"prod/internal/adapters/controller/api/validator"
	"prod/internal/adapters/database/postgres"
	"prod/internal/domain/common/errorz"
	"prod/internal/domain/dto"
	"prod/internal/domain/entity"
	"prod/internal/domain/service"
	"strconv"
)

type PromoService interface {
	GetFeed(ctx context.Context, user *entity.User, dto dto.PromoFeedRequest) ([]dto.PromoForUser, int64, error)
	GetByIdUser(ctx context.Context, promoID, userID string) (dto.PromoForUser, error)
	GetHistory(ctx context.Context, userID string, limit, offset int) ([]dto.PromoForUser, int64, error)
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
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if requestDTO.Limit == 0 {
		requestDTO.Limit = 10
	}

	user := c.Locals("user").(*entity.User)

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Пользователь не авторизован.",
		})
	}

	promos, total, err := h.PromoService.GetFeed(c.Context(), user, requestDTO)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	c.Append("X-Total-Count", strconv.FormatInt(total, 10))

	return c.Status(fiber.StatusOK).JSON(promos)
}

func (h UserPromoHandler) GetPromoByID(c fiber.Ctx) error {
	var requestDTO dto.PromoGetByID

	user := c.Locals("user").(*entity.User)

	if err := c.Bind().URI(&requestDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(requestDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Пользователь не авторизован.",
		})
	}

	promo, err := h.PromoService.GetByIdUser(c.Context(), requestDTO.ID, user.ID)

	if err != nil {
		if errors.Is(err, errorz.NotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Промо не найдено.",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(promo)
}

func (h UserPromoHandler) GetHistory(c fiber.Ctx) error {
	var requestDTO dto.PromoHistory
	user := c.Locals("user").(*entity.User)

	if err := c.Bind().Query(&requestDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if requestDTO.Limit == 0 {
		requestDTO.Limit = 10
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Пользователь не авторизован.",
		})
	}

	promos, total, err := h.PromoService.GetHistory(c.Context(), user.ID, requestDTO.Limit, requestDTO.Offset)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	c.Append("X-Total-Count", strconv.FormatInt(total, 10))

	return c.Status(fiber.StatusOK).JSON(promos)
}

func (h UserPromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	userGroup := router.Group("/user")
	userGroup.Get("/feed", h.GetFeed, middleware)
	userGroup.Get("/promo/history", h.GetHistory, middleware)
	userGroup.Get("/promo/:id", h.GetPromoByID, middleware)
}
