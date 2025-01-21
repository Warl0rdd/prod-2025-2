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
	GetFeed(ctx context.Context, user *entity.User, dto dto.PromoFeedRequest) ([]dto.PromoFeed, int64, error)
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

	return c.JSON(promos)
}

func (h UserPromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	userGroup := router.Group("/user")
	userGroup.Get("/feed", h.GetFeed, middleware)
}
