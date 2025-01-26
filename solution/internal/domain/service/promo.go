package service

import (
	"context"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"strings"
	"time"
)

type promoStorage interface {
	Create(ctx context.Context, promo entity.Promo) (*entity.Promo, error)
	GetByID(ctx context.Context, id string) (*entity.Promo, error)
	Update(ctx context.Context, fiberCtx fiber.Ctx, promo dto.PromoUpdate, id string) (*entity.Promo, error)
	GetWithPagination(ctx context.Context, limit, offset int, sortBy, companyId string, countries []countries.CountryCode) ([]entity.Promo, int64, error)
	GetFeed(ctx context.Context, age, limit, offset int, country countries.CountryCode, category *string, active, userID string) ([]dto.PromoForUser, int64, error)
	GetByIdUser(ctx context.Context, promoID, userID string) (dto.PromoForUser, error)
	GetHistory(ctx context.Context, userID string, limit, offset int) ([]dto.PromoForUser, int64, error)
	GetStats(ctx context.Context, promoID, companyID string) (dto.PromoStatsResponse, error)
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
			//activeFrom, timeError = time.Parse("2006-01-02 15:04:05", promoDTO.ActiveFrom)
			//if timeError != nil {
			//	return nil, timeError
			//}

			return nil, errorz.BadRequest
		}
	} else {
		activeFrom = time.Unix(0, 0)
	}
	if promoDTO.ActiveUntil != "" {
		activeUntil, timeError = time.Parse("2006-01-02", promoDTO.ActiveUntil)
		if timeError != nil {
			//activeUntil, timeError = time.Parse("2006-01-02 15:04:05", promoDTO.ActiveUntil)
			//if timeError != nil {
			//	return nil, timeError
			//}

			return nil, errorz.BadRequest
		}
	} else {
		activeUntil = time.Unix(8210266876, 0)
	}

	var categories []entity.Category
	var promoUniques []entity.PromoUnique

	if promoDTO.Target != nil {
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
	}

	company := fiberCTX.Locals("business").(*entity.Business)

	promo := entity.Promo{
		CompanyID:   company.ID,
		Active:      promoDTO.Active,
		ActiveFrom:  activeFrom,
		ActiveUntil: activeUntil,
		Description: promoDTO.Description,
		ImageURL:    promoDTO.ImageURL,
		MaxCount:    promoDTO.MaxCount,
		Mode:        promoDTO.Mode,
		PromoCommon: promoDTO.PromoCommon,
		PromoUnique: promoUniques,
		Categories:  categories,
	}

	if promoDTO.Target != nil {
		promo.Country = countries.ByName(strings.ToUpper(promoDTO.Target.Country))
		promo.AgeUntil = promoDTO.Target.AgeUntil
		if promoDTO.Target.AgeUntil == 0 {
			promo.AgeUntil = 1000
		}
		promo.AgeFrom = promoDTO.Target.AgeFrom
	} else {
		promo.AgeFrom = 0
		promo.AgeUntil = 1000
	}

	if activeFrom.After(time.Now()) || activeUntil.Before(time.Now()) {
		promo.Active = false
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

func (s *promoService) Update(ctx context.Context, fiberCtx fiber.Ctx, dto dto.PromoUpdate, id string) (*entity.Promo, error) {
	return s.promoStorage.Update(ctx, fiberCtx, dto, id)
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

func (s *promoService) GetStats(ctx context.Context, promoID, companyID string) (dto.PromoStatsResponse, error) {
	return s.promoStorage.GetStats(ctx, promoID, companyID)
}
