package b2b

import (
	"context"
	"errors"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"slices"
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
	Update(ctx context.Context, fiberCtx fiber.Ctx, dto dto.PromoUpdate, id string) (*entity.Promo, error)
	GetStats(ctx context.Context, promoID, companyID string) (dto.PromoStatsResponse, error)
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

	if len(promoDTO.Description) < 10 || len(promoDTO.Description) > 300 {
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

	if len(promoDTO.Target.Country) > 2 {
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

	if promoDTO.Target.AgeFrom != 0 && (promoDTO.Target.AgeUntil != 0 && promoDTO.Target.AgeUntil < promoDTO.Target.AgeFrom) {
		logger.Log.Error("age until")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	promo, err := h.promoService.Create(c.Context(), c, promoDTO)
	if err != nil {
		if errors.Is(err, errorz.BadRequest) {
			return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: err.Error(),
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: err.Error(),
			})
		}
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

	if promoRequestDTO.SortBy != "active_from" && promoRequestDTO.SortBy != "active_until" && promoRequestDTO.SortBy != "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
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
		promoDTO.Countries = append(promoDTO.Countries, countries.ByName(country))
	}

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
			if category.Name == "" {
				continue
			}
			categories = append(categories, category.Name)
		}

		for _, promoUnique := range promo.PromoUnique {
			if promoUnique.Body == "" {
				continue
			}
			promoUniques = append(promoUniques, promoUnique.Body)
		}

		promoDTOs = append(promoDTOs, dto.PromoDTO{
			PromoID:     promo.PromoID,
			CompanyID:   promo.CompanyID,
			CompanyName: company.Name,
			Target: dto.Target{
				AgeFrom:    promo.AgeFrom,
				AgeUntil:   promo.AgeUntil,
				Country:    strings.ToLower(promo.Country.Alpha2()),
				Categories: categories,
			},
			Active:      promo.Active,
			ActiveFrom:  promo.ActiveFrom.Format("2006-01-02"),
			ActiveUntil: promo.ActiveUntil.Format("2006-01-02"),
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
			Country:    strings.ToLower(promo.Country.Alpha2()),
			Categories: categories,
		},
		Active:      promo.Active,
		ActiveFrom:  promo.ActiveFrom.Format("2006-01-02"),
		ActiveUntil: promo.ActiveUntil.Format("2006-01-02"),
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
	var promoDTO dto.PromoUpdate

	business := c.Locals("business").(*entity.Business)

	if err := c.Bind().Body(&promoDTO); err != nil {
		logger.Log.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if err := c.Bind().URI(&params); err != nil {
		logger.Log.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if (promoDTO.Description != nil) && (len(*promoDTO.Description) < 10 || len(*promoDTO.Description) > 300) {
		logger.Log.Error("desc len")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if (promoDTO.Mode != nil) && (*promoDTO.Mode != "COMMON") && (*promoDTO.Mode != "UNIQUE") {
		logger.Log.Error("mode")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if (promoDTO.Mode != nil) && (promoDTO.MaxCount != nil) && (promoDTO.PromoUnique == nil || (*promoDTO.MaxCount != 1)) && *promoDTO.Mode == "UNIQUE" {
		logger.Log.Error("unique max count")
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if promoDTO.Mode != nil {
		mode := *promoDTO.Mode
		if (mode == "COMMON" && promoDTO.PromoUnique != nil) ||
			(mode == "UNIQUE" && promoDTO.PromoCommon != nil) {
			logger.Log.Error("common/unique fields conflict")
			return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка в данных запроса.",
			})
		}
	}

	if promoDTO.Target != nil {
		if len(promoDTO.Target.Country) > 2 {
			logger.Log.Error("country len")
			return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка в данных запроса.",
			})
		}

		if countryCode := countries.ByName(strings.ToUpper(promoDTO.Target.Country)); countryCode == countries.Unknown && promoDTO.Target.Country != "" {
			logger.Log.Error("country not found")
			return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка в данных запроса.",
			})
		}

		if (promoDTO.Target.AgeFrom != 0) && (promoDTO.Target.AgeUntil != 0) && (promoDTO.Target.AgeUntil < promoDTO.Target.AgeFrom) {
			logger.Log.Error("age until")
			return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка в данных запроса.",
			})
		}

		if slices.Contains(promoDTO.Target.Categories, "") || slices.Contains(promoDTO.PromoUnique, "") {
			logger.Log.Error("empty category")
			return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: "Ошибка в данных запроса.",
			})
		}
	}

	promo, err := h.promoService.Update(c.Context(), c, promoDTO, params.ID)

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

	if errors.Is(err, errorz.BadRequest) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	var categories []string
	var promoUniques []string

	for _, category := range promo.Categories {
		categories = append(categories, category.Name)
	}

	for _, promoUnique := range promo.PromoUnique {
		promoUniques = append(promoUniques, promoUnique.Body)
	}

	promoReturn := dto.PromoDTO{
		PromoID:     promo.PromoID,
		CompanyID:   promo.CompanyID,
		CompanyName: business.Name,
		Target: dto.Target{
			AgeFrom:    promo.AgeFrom,
			AgeUntil:   promo.AgeUntil,
			Categories: categories,
		},
		Active:      promo.Active,
		ActiveFrom:  promo.ActiveFrom.Format("2006-01-02"),
		ActiveUntil: promo.ActiveUntil.Format("2006-01-02"),
		Description: promo.Description,
		ImageURL:    promo.ImageURL,
		MaxCount:    promo.MaxCount,
		Mode:        promo.Mode,
		LikeCount:   promo.LikeCount,
		UsedCount:   promo.UsedCount,
		PromoCommon: promo.PromoCommon,
		PromoUnique: promoUniques,
	}

	if promo.Country != 0 {
		promoReturn.Target.Country = strings.ToLower(promo.Country.Alpha2())
	}

	return c.Status(fiber.StatusOK).JSON(promoReturn)
}

func (h PromoHandler) stats(c fiber.Ctx) error {
	var requestDTO dto.PromoStats
	business := c.Locals("business").(*entity.Business)

	if business == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Пользователь не авторизован.",
		})
	}

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

	promoByID, err := h.promoService.GetByID(c.Context(), requestDTO.Id)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if promoByID.CompanyID != business.ID {
		return c.Status(fiber.StatusForbidden).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Промокод не принадлежит этой компании.",
		})
	}

	promos, statsErr := h.promoService.GetStats(c.Context(), requestDTO.Id, business.ID)

	if statsErr != nil {
		if errors.Is(err, errorz.NotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.HTTPResponse{
				Status:  "error",
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(promos)
}

func (h PromoHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	promoGroup := router.Group("/business")
	promoGroup.Post("/promo", h.create, middleware)
	promoGroup.Get("/promo", h.getWithPagination, middleware)
	promoGroup.Get("/promo/:id", h.getByID, middleware)
	promoGroup.Patch("/promo/:id", h.update, middleware)
	promoGroup.Get("/promo/:id/stat", h.stats, middleware)
}
