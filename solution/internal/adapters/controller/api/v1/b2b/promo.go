package b2b

import (
	"context"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"solution/cmd/app"
	"solution/internal/adapters/controller/api/validator"
	"solution/internal/adapters/database/postgres"
	"solution/internal/adapters/logger"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"strconv"
	"strings"
)

type PromoService interface {
	Create(ctx context.Context, fiberCTX fiber.Ctx, promoDTO dto.PromoCreate) (*entity.Promo, error)
	GetByID(ctx context.Context, uuid string) (*entity.Promo, error)
	GetWithPagination(ctx context.Context, dto dto.PromoGetWithPagination) ([]entity.Promo, int64, error)
	Update(ctx context.Context, promo *entity.Promo) (*entity.Promo, error)
}

type PromoHandler struct {
	promoService PromoService
	validator    *validator.Validator
}

func NewPromoHandler(app *app.App) *PromoHandler {
	promoStorage := postgres.NewPromoStorage(app.DB)
	businessStorage := postgres.NewBusinessStorage(app.DB)

	return &PromoHandler{
		promoService: service.NewPromoService(promoStorage, businessStorage),
		validator:    app.Validator,
	}
}

func (h PromoHandler) create(c fiber.Ctx) error {
	var promoDTO dto.PromoCreate

	if c.Locals("business") == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Пользователь не авторизован.",
		})
	}

	if err := c.Bind().Body(&promoDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(promoDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoDTO.Mode != "COMMON" && promoDTO.Mode != "UNIQUE" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoDTO.Mode == "UNIQUE" && (promoDTO.PromoUnique == nil || promoDTO.MaxCount != 1) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if countryCode := countries.ByName(strings.ToUpper(promoDTO.Target.Country)); countryCode == countries.Unknown {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promo, err := h.promoService.Create(c.Context(), c, promoDTO)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.PromoCreateResponse{
		ID: promo.PromoID,
	})
}

func (h PromoHandler) getWithPagination(c fiber.Ctx) error {
	var promoRequestDTO dto.PromoGetWithPaginationRequest

	if err := c.Bind().Header(&promoRequestDTO); err != nil {
		logger.Log.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoRequestDTO.Limit == 0 {
		promoRequestDTO.Limit = 10
	}

	if promoRequestDTO.SortBy != "active_from" && promoRequestDTO.SortBy != "active_until" {
		logger.Log.Error(promoRequestDTO.SortBy)
		logger.Log.Error("active_from err")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promoDTO := dto.PromoGetWithPagination{
		Limit:  promoRequestDTO.Limit,
		Offset: promoRequestDTO.Offset,
		SortBy: promoRequestDTO.SortBy,
	}

	for _, country := range promoRequestDTO.Countries {
		promoDTO.Countries = append(promoDTO.Countries, countries.ByName(strings.ToUpper(country)))
	}

	promos, total, err := h.promoService.GetWithPagination(c.Context(), promoDTO)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Status:  "error",
			Message: err.Error(),
		})
	}

	c.Append("X-Total-Count", strconv.FormatInt(total, 10))

	return c.Status(fiber.StatusOK).JSON(promos)
}

func (h PromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	promoGroup := router.Group("/business")
	promoGroup.Post("/promo", h.create, middleware)
	promoGroup.Get("/promo", h.getWithPagination, middleware)
}
