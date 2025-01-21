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
)

type ActionsService interface {
	AddLike(ctx context.Context, userID, promoID string) error
}

type ActionsHandler struct {
	actionsService ActionsService
	validator      *validator.Validator
}

func NewActionsHandler(app *app.App) *ActionsHandler {
	actionsStorage := postgres.NewActionsStorage(app.DB)

	return &ActionsHandler{
		actionsService: service.NewActionsService(actionsStorage),
		validator:      app.Validator,
	}
}

func (h ActionsHandler) AddLike(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)
	var likeDTO dto.AddLike

	if err := c.Bind().URI(&likeDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(likeDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	err := h.actionsService.AddLike(c.Context(), user.ID, likeDTO.PromoID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка сервера.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.HTTPResponse{
		Status: "ok",
	})
}

func (h ActionsHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	actionsGroup := router.Group("/user/promo")

	actionsGroup.Post("/:id/like", h.AddLike, middleware)
}
