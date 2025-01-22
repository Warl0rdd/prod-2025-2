package b2b

import (
	"context"
	"errors"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"solution/cmd/app"
	"solution/internal/adapters/controller/api/validator"
	"solution/internal/adapters/database/postgres"
	"solution/internal/adapters/logger"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"strconv"
	"strings"
)

type PromoService interface {
	Create(ctx context.Context, fiberCTX fiber.Ctx, promoDTO dto.PromoCreate) (*entity.Promo, error)
	GetByID(ctx context.Context, uuid string) (*entity.Promo, error)
	GetWithPagination(ctx context.Context, companyId string, dto dto.PromoGetWithPagination) ([]entity.Promo, int64, error)
	Update(ctx context.Context, fiberCtx fiber.Ctx, dto dto.PromoCreate, id string) (*entity.Promo, error)
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

// Создание промо
func (h PromoHandler) create(c fiber.Ctx) error {
	var promoDTO dto.PromoCreate

	if err := c.Bind().Body(&promoDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(promoDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if (promoDTO.Mode != "COMMON") && (promoDTO.Mode != "UNIQUE") {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoDTO.Mode == "UNIQUE" && (promoDTO.PromoUnique == nil || promoDTO.MaxCount != 1) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if (promoDTO.Mode == "COMMON" && promoDTO.PromoUnique != nil) || (promoDTO.Mode == "UNIQUE" && promoDTO.PromoCommon != "") {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if countryCode := countries.ByName(strings.ToUpper(promoDTO.Target.Country)); countryCode == countries.Unknown && promoDTO.Target.Country != "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promo, err := h.promoService.Create(c.Context(), c, promoDTO)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.PromoCreateResponse{
		ID: promo.PromoID,
	})
}

// Получение промо с пагинацией
func (h PromoHandler) getWithPagination(c fiber.Ctx) error {
	var promoRequestDTO dto.PromoGetWithPaginationRequest

	company := c.Locals("business").(*entity.Business)

	if err := c.Bind().Query(&promoRequestDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoRequestDTO.Limit == 0 {
		promoRequestDTO.Limit = 10
	}

	if promoRequestDTO.SortBy != "active_from" && promoRequestDTO.SortBy != "active_until" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	logger.Log.Info(promoRequestDTO)

	promoDTO := dto.PromoGetWithPagination{
		Limit:  promoRequestDTO.Limit,
		Offset: promoRequestDTO.Offset,
		SortBy: promoRequestDTO.SortBy,
	}

	for _, country := range promoRequestDTO.Countries {
		promoDTO.Countries = append(promoDTO.Countries, countries.ByName(country))
	}

	logger.Log.Info(promoDTO)

	promos, total, err := h.promoService.GetWithPagination(c.Context(), company.ID, promoDTO)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	var promoDTOs []dto.PromoDTO

	for _, promo := range promos {

		var categories, promoUniques []string

		for _, category := range promo.Categories {
			categories = append(categories, category.Name)
		}

		for _, promoUnique := range promo.PromoUnique {
			promoUniques = append(promoUniques, promoUnique.Body)
		}

		promoDTOs = append(promoDTOs, dto.PromoDTO{
			PromoID:     promo.PromoID,
			CompanyID:   promo.CompanyID,
			CompanyName: company.Name,
			Target: dto.Target{
				AgeFrom:    promo.AgeFrom,
				AgeUntil:   promo.AgeUntil,
				Country:    promo.Country.Alpha2(),
				Categories: categories,
			},
			Active:      promo.Active,
			ActiveFrom:  promo.ActiveFrom,
			ActiveUntil: promo.ActiveUntil,
			Description: promo.Description,
			ImageURL:    promo.ImageURL,
			MaxCount:    promo.MaxCount,
			Mode:        promo.Mode,
			LikeCount:   promo.LikeCount,
			UsedCount:   promo.UsedCount,
			PromoCommon: promo.PromoCommon,
			PromoUnique: promoUniques,
		})
	}

	c.Append("X-Total-Count", strconv.FormatInt(total, 10))

	return c.Status(fiber.StatusOK).JSON(promoDTOs)
}

// Получение промо по ID
func (h PromoHandler) getByID(c fiber.Ctx) error {
	var promoIdDTO dto.PromoGetByID
	business := c.Locals("business").(*entity.Business)

	if err := c.Bind().URI(&promoIdDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(promoIdDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promo, err := h.promoService.GetByID(c.Context(), promoIdDTO.ID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if promo.CompanyID != business.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Промокод не принадлежит этой компании.",
		})
	}

	var categories, promoUniques []string

	for _, category := range promo.Categories {
		categories = append(categories, category.Name)
	}

	for _, promoUnique := range promo.PromoUnique {
		promoUniques = append(promoUniques, promoUnique.Body)
	}

	return c.Status(fiber.StatusOK).JSON(dto.PromoDTO{
		PromoID:     promo.PromoID,
		CompanyID:   promo.CompanyID,
		CompanyName: business.Name,
		Target: dto.Target{
			AgeFrom:    promo.AgeFrom,
			AgeUntil:   promo.AgeUntil,
			Country:    promo.Country.Alpha2(),
			Categories: categories,
		},
		Active:      promo.Active,
		ActiveFrom:  promo.ActiveFrom,
		ActiveUntil: promo.ActiveUntil,
		Description: promo.Description,
		ImageURL:    promo.ImageURL,
		MaxCount:    promo.MaxCount,
		Mode:        promo.Mode,
		LikeCount:   promo.LikeCount,
		UsedCount:   promo.UsedCount,
		PromoCommon: promo.PromoCommon,
		PromoUnique: promoUniques,
	})
}

// Обновление промо
func (h PromoHandler) update(c fiber.Ctx) error {
	type Params struct {
		ID string `uri:"id"`
	}

	var params Params
	var promoDTO dto.PromoCreate

	if err := c.Bind().Body(&promoDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if err := c.Bind().URI(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	newPromo, err := h.promoService.Update(c.Context(), c, promoDTO, params.ID)

	if errors.Is(err, errorz.Forbidden) {
		return c.Status(fiber.StatusForbidden).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Промокод не принадлежит этой компании.",
		})
	}

	if errors.Is(err, errorz.NotFound) {
		return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Промокод не найден.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(newPromo)
}

func (h PromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	promoGroup := router.Group("/business")
	promoGroup.Post("/promo", h.create, middleware)
	promoGroup.Get("/promo", h.getWithPagination, middleware)
	promoGroup.Get("/promo/:id", h.getByID, middleware)
	promoGroup.Patch("/promo/:id", h.update, middleware)
}
