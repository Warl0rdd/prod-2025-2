package b2c

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v3"
	"prod/cmd/app"
	"prod/internal/adapters/controller/api/validator"
	"prod/internal/adapters/database/postgres"
	"prod/internal/adapters/database/redis"
	"prod/internal/adapters/logger"
	"prod/internal/domain/common/errorz"
	"prod/internal/domain/dto"
	"prod/internal/domain/entity"
	"prod/internal/domain/service"
	"strconv"
)

type ActionsService interface {
	AddLike(ctx context.Context, userID, promoID string) error
	DeleteLike(ctx context.Context, userID, promoID string) error
	AddComment(ctx context.Context, userID, promoID, text string) (string, error)
	GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, int64, error)
	GetCommentById(ctx context.Context, commentID, promoID string) (dto.Comment, error)
	UpdateComment(ctx context.Context, promoID, commentID, userID, text string) (dto.Comment, error)
	DeleteComment(ctx context.Context, promoID, commentID, userID string) error
	Activate(ctx context.Context, user *entity.User, promoID string) (string, error)
}

type ActionsHandler struct {
	actionsService ActionsService
	validator      *validator.Validator
}

func NewActionsHandler(app *app.App) *ActionsHandler {
	actionsStorage := postgres.NewActionsStorage(app.DB)
	activationStorage := postgres.NewActivationStorage(app.DB)
	activationRedisStorage := redis.NewActivationStorage(app.Redis)

	return &ActionsHandler{
		actionsService: service.NewActionsService(actionsStorage, activationStorage, activationRedisStorage),
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

	id, err := h.actionsService.AddComment(c.Context(), user.ID, commentDTO.PromoID, commentDTO.Text)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка сервера.",
		})
	}

	comment, errGet := h.actionsService.GetCommentById(c.Context(), id, commentDTO.PromoID)

	if errGet != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка сервера.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
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

	if getCommentsDTO.Limit == 0 {
		getCommentsDTO.Limit = 10
	}

	comments, total, err := h.actionsService.GetComments(ctx.Context(), getCommentsDTO.ID, getCommentsDTO.Limit, getCommentsDTO.Offset)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка сервера.",
		})
	}

	ctx.Append("X-Total-Count", strconv.FormatInt(total, 10))

	return ctx.Status(fiber.StatusOK).JSON(comments)
}

func (h ActionsHandler) getCommentById(ctx fiber.Ctx) error {
	var getCommentDTO dto.GetCommentById

	if err := ctx.Bind().URI(&getCommentDTO); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(getCommentDTO); errValidate != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	comment, err := h.actionsService.GetCommentById(ctx.Context(), getCommentDTO.CommentID, getCommentDTO.ID)

	if err != nil {
		if errors.Is(err, errorz.NotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Комментарий не найден.",
			})
		} else {
			return ctx.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка сервера.",
			})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(comment)
}

func (h ActionsHandler) updateComment(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)
	var commentDTO dto.UpdateComment

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

	comment, err := h.actionsService.UpdateComment(c.Context(), commentDTO.ID, commentDTO.CommentID, user.ID, commentDTO.Text)

	if err != nil {
		if errors.Is(err, errorz.Forbidden) {
			return c.Status(fiber.StatusForbidden).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Недостаточно прав.",
			})
		} else if errors.Is(err, errorz.NotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Комментарий не найден.",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка сервера.",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(comment)
}

func (h ActionsHandler) deleteComment(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)
	var commentDTO dto.DeleteCommentById

	if err := c.Bind().URI(&commentDTO); err != nil {
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

	err := h.actionsService.DeleteComment(c.Context(), commentDTO.ID, commentDTO.CommentID, user.ID)

	if err != nil {
		if errors.Is(err, errorz.Forbidden) {
			return c.Status(fiber.StatusForbidden).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Недостаточно прав.",
			})
		} else if errors.Is(err, errorz.NotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Комментарий не найден.",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка сервера.",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(dto.HTTPResponse{
		Status: "ok",
	})
}

func (h ActionsHandler) activate(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)
	var activateDTO dto.Activate

	if err := c.Bind().URI(&activateDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if err := h.validator.ValidateData(activateDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promo, err := h.actionsService.Activate(c.Context(), user, activateDTO.ID)

	if err != nil {
		if errors.Is(err, errorz.Forbidden) {
			return c.Status(fiber.StatusForbidden).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Доступ запрещен.",
			})
		} else if errors.Is(err, errorz.NotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Промо не найдено.",
			})

		} else {
			logger.Log.Error(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка сервера.",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(dto.ActivateResponse{Promo: promo})
}

func (h ActionsHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	actionsGroup := router.Group("/user/promo")

	actionsGroup.Post("/:id/like", h.addLike, middleware)
	actionsGroup.Delete("/:id/like", h.deleteLike, middleware)
	actionsGroup.Post("/:id/comments", h.addComment, middleware)
	actionsGroup.Get("/:id/comments", h.getComments, middleware)
	actionsGroup.Get("/:id/comments/:comment_id", h.getCommentById, middleware)
	actionsGroup.Put("/:id/comments/:comment_id", h.updateComment, middleware)
	actionsGroup.Delete("/:id/comments/:comment_id", h.deleteComment, middleware)
	actionsGroup.Post("/:id/activate", h.activate, middleware)
}
