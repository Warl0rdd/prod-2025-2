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
	DeleteLike(ctx context.Context, userID, promoID string) error
	AddComment(ctx context.Context, userID, promoID, text string) error
	GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, error)
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

func (h ActionsHandler) addLike(c fiber.Ctx) error {
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

func (h ActionsHandler) deleteLike(c fiber.Ctx) error {
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

	err := h.actionsService.DeleteLike(c.Context(), user.ID, likeDTO.PromoID)

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

func (h ActionsHandler) addComment(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)
	var commentDTO dto.AddComment

	if err := c.Bind().URI(&commentDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if err := c.Bind().Body(&commentDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(commentDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	err := h.actionsService.AddComment(c.Context(), user.ID, commentDTO.PromoID, commentDTO.Text)

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

func (h ActionsHandler) getComments(ctx fiber.Ctx) error {
	var getCommentsDTO dto.GetComments

	if err := ctx.Bind().URI(&getCommentsDTO); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if err := ctx.Bind().Query(&getCommentsDTO); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(getCommentsDTO); errValidate != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	comments, err := h.actionsService.GetComments(ctx.Context(), getCommentsDTO.ID, getCommentsDTO.Limit, getCommentsDTO.Offset)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка сервера.",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(comments)
}

func (h ActionsHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	actionsGroup := router.Group("/user/promo")

	actionsGroup.Post("/:id/like", h.addLike, middleware)
	actionsGroup.Delete("/:id/like", h.deleteLike, middleware)
	actionsGroup.Post("/:id/comments", h.addComment, middleware)
	actionsGroup.Get("/:id/comments", h.getComments, middleware)
}
