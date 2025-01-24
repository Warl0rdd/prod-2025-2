package service

import (
	"context"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"strings"
	"time"
)

type promoStorage interface {
	Create(ctx context.Context, promo entity.Promo) (*entity.Promo, error)
	GetByID(ctx context.Context, id string) (*entity.Promo, error)
	Update(ctx context.Context, fiberCtx fiber.Ctx, promo *entity.Promo, id string) (*entity.Promo, error)
	GetWithPagination(ctx context.Context, limit, offset int, sortBy, companyId string, countries []countries.CountryCode) ([]entity.Promo, int64, error)
	GetFeed(ctx context.Context, age, limit, offset int, country countries.CountryCode, category, active, userID string) ([]dto.PromoForUser, int64, error)
	GetByIdUser(ctx context.Context, promoID, userID string) (dto.PromoForUser, error)
	GetHistory(ctx context.Context, userID string, limit, offset int) ([]dto.PromoForUser, int64, error)
}

type promoService struct {
	promoStorage    promoStorage
	businessStorage businessStorage
}

func NewPromoService(promoStorage promoStorage, businessStorage businessStorage) *promoService {
	return &promoService{
		promoStorage:    promoStorage,
		businessStorage: businessStorage,
	}
}

func (s *promoService) Create(ctx context.Context, fiberCTX fiber.Ctx, promoDTO dto.PromoCreate) (*entity.Promo, error) {
	var activeFrom, activeUntil time.Time
	var timeError error
	if promoDTO.ActiveFrom != "" {
		activeFrom, timeError = time.Parse("2006-01-02", promoDTO.ActiveFrom)
		if timeError != nil {
			return nil, timeError
		}
	}
	if promoDTO.ActiveUntil != "" {
		activeUntil, timeError = time.Parse("2006-01-02", promoDTO.ActiveUntil)
		if timeError != nil {
			return nil, timeError
		}
	}

	var categories []entity.Category
	var promoUniques []entity.PromoUnique
	for _, category := range promoDTO.Target.Categories {
		categories = append(categories, entity.Category{
			Name: category,
		})
	}
	for _, promoUnique := range promoDTO.PromoUnique {
		promoUniques = append(promoUniques, entity.PromoUnique{
			Body: promoUnique,
		})
	}

	company := fiberCTX.Locals("business").(*entity.Business)

	promo := entity.Promo{
		CompanyID:   company.ID,
		Active:      true,
		ActiveFrom:  activeFrom,
		ActiveUntil: activeUntil,
		Description: promoDTO.Description,
		ImageURL:    promoDTO.ImageURL,
		MaxCount:    promoDTO.MaxCount,
		Mode:        promoDTO.Mode,
		PromoCommon: promoDTO.PromoCommon,
		PromoUnique: promoUniques,
		AgeFrom:     promoDTO.Target.AgeFrom,
		AgeUntil:    promoDTO.Target.AgeUntil,
		Country:     countries.ByName(strings.ToUpper(promoDTO.Target.Country)),
		Categories:  categories,
	}

	company.Promos = append(company.Promos, promo)
	_, err := s.businessStorage.Update(ctx, company)
	if err != nil {
		return nil, err
	}

	return s.promoStorage.Create(ctx, promo)
}

func (s *promoService) GetByID(ctx context.Context, id string) (*entity.Promo, error) {
	return s.promoStorage.GetByID(ctx, id)
}

func (s *promoService) GetWithPagination(ctx context.Context, companyId string, dto dto.PromoGetWithPagination) ([]entity.Promo, int64, error) {
	return s.promoStorage.GetWithPagination(ctx, dto.Limit, dto.Offset, dto.SortBy, companyId, dto.Countries)
}

func (s *promoService) Update(ctx context.Context, fiberCtx fiber.Ctx, dto dto.PromoCreate, id string) (*entity.Promo, error) {
	var activeFrom, activeUntil time.Time
	var timeError error
	if dto.ActiveFrom != "" {
		activeFrom, timeError = time.Parse("2006-01-02", dto.ActiveFrom)
		if timeError != nil {
			return nil, timeError
		}
	}
	if dto.ActiveUntil != "" {
		activeUntil, timeError = time.Parse("2006-01-02", dto.ActiveUntil)
		if timeError != nil {
			return nil, timeError
		}
	}

	var categories []entity.Category
	var promoUniques []entity.PromoUnique
	for _, category := range dto.Target.Categories {
		categories = append(categories, entity.Category{
			Name: category,
		})
	}
	for _, promoUnique := range dto.PromoUnique {
		promoUniques = append(promoUniques, entity.PromoUnique{
			Body: promoUnique,
		})
	}

	promo := entity.Promo{
		AgeFrom:     dto.Target.AgeFrom,
		AgeUntil:    dto.Target.AgeUntil,
		Country:     countries.ByName(strings.ToUpper(dto.Target.Country)),
		Categories:  categories,
		Active:      true,
		ActiveFrom:  activeFrom,
		ActiveUntil: activeUntil,
		Description: dto.Description,
		ImageURL:    dto.ImageURL,
		MaxCount:    dto.MaxCount,
		Mode:        dto.Mode,
		PromoCommon: dto.PromoCommon,
		PromoUnique: promoUniques,
	}

	return s.promoStorage.Update(ctx, fiberCtx, &promo, id)
}

func (s *promoService) GetFeed(ctx context.Context, user *entity.User, dto dto.PromoFeedRequest) ([]dto.PromoForUser, int64, error) {
	return s.promoStorage.GetFeed(ctx, user.Age, dto.Limit, dto.Offset, user.Country, dto.Category, dto.Active, user.ID)
}

func (s *promoService) GetByIdUser(ctx context.Context, promoID, userID string) (dto.PromoForUser, error) {
	return s.promoStorage.GetByIdUser(ctx, promoID, userID)
}

func (s *promoService) GetHistory(ctx context.Context, userID string, limit, offset int) ([]dto.PromoForUser, int64, error) {
	return s.promoStorage.GetHistory(ctx, userID, limit, offset)
}
