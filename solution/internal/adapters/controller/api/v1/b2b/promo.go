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
	"strings"
)

type PromoService interface {
	Create(ctx context.Context, registerReq dto.PromoCreate) (*entity.Promo, error)
	GetByID(ctx context.Context, uuid string) (*entity.Promo, error)
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

	companyID := c.Locals("business").(*entity.Business).ID
	promoDTO.CompanyID = companyID

	if err := c.Bind().Body(&promoDTO); err != nil {
		logger.Log.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(promoDTO); errValidate != nil {
		logger.Log.Error(errValidate)
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoDTO.Mode != "COMMON" && promoDTO.Mode != "UNIQUE" {
		logger.Log.Error("COMMON ERROR")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoDTO.Mode == "UNIQUE" && (promoDTO.PromoUnique == nil || promoDTO.MaxCount != 1) {
		logger.Log.Error("UNIQUE ERROR")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if countryCode := countries.ByName(strings.ToUpper(promoDTO.Target.Country)); countryCode == countries.Unknown {
		logger.Log.Error("COUNTRY ERROR")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promo, err := h.promoService.Create(c.Context(), promoDTO)
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

func (h PromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	promoGroup := router.Group("/business")
	promoGroup.Post("/promo", h.create, middleware)
}
