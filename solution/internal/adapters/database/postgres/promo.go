package postgres

import (
	"context"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"slices"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/entity"
)

// promoStorage is a struct that contains a pointer to a gorm.DB instance to interact with promo repository.
type promoStorage struct {
	db *gorm.DB
}

// NewPromoStorage is a function that returns a new instance of promoStorage.
func NewPromoStorage(db *gorm.DB) *promoStorage {
	return &promoStorage{db: db}
}

// Create is a method to create a new Promo in database.
func (s *promoStorage) Create(ctx context.Context, promo entity.Promo) (*entity.Promo, error) {
	// Insert a promo (parent)'s entity
	insertPromoQuery := s.db.WithContext(ctx).Raw(
		"INSERT INTO promos (company_id, created_at, updated_at, active_from, active_until, description, image_url, max_count, mode, promo_common, age_from, age_unti, country) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING promo_id;",
		promo.CompanyID, promo.CreatedAt, promo.UpdatedAt, promo.ActiveFrom, promo.ActiveUntil, promo.Description, promo.ImageURL, promo.MaxCount, promo.Mode, promo.PromoCommon, promo.AgeFrom, promo.AgeUntil, promo.Country).Scan(&promo.PromoID)
	if err := insertPromoQuery.Error; err != nil {
		return nil, err
	}

	// Insert categories
	for _, category := range promo.Categories {
		insertCategoryQuery := s.db.WithContext(ctx).Exec("INSERT INTO categories (promo_id, name) VALUES (?, ?);", promo.PromoID, category.Name)
		if err := insertCategoryQuery.Error; err != nil {
			return nil, err
		}
	}

	// Insert promo_uniques
	for _, promoUnique := range promo.PromoUnique {
		insertPromoUniqueQuery := s.db.WithContext(ctx).Exec("INSERT INTO promo_uniques (promo_id, body, activated) VALUES (?, ?, ?);", promo.PromoID, promoUnique.Body, promoUnique.Activated)
		if err := insertPromoUniqueQuery.Error; err != nil {
			return nil, err
		}
	}

	return &promo, nil
}

// GetByID is a method that returns an error and a pointer to a Promo instance by id.
func (s *promoStorage) GetByID(ctx context.Context, id string) (*entity.Promo, error) {
	var promo entity.Promo

	query := `
		SELECT 
			p.*, 
			c.category_id, c.name, 
			pu.promo_unique_id, pu.body, pu.activated
		FROM 
			promos p
		LEFT JOIN 
			categories c ON p.promo_id = c.promo_id
		LEFT JOIN 
			promo_uniques pu ON p.promo_id = pu.promo_id
		WHERE 
			p.promo_id = ?
	`

	// Временные структуры для хранения данных
	type result struct {
		entity.Promo
		CategoryID    string
		Name          string
		PromoUniqueID string
		Body          string
		Activated     bool
	}

	var results []result

	// Выполнение запроса
	if err := s.db.Raw(query, id).Scan(&results).Error; err != nil {
		return nil, err
	}

	// Обработка результатов
	if len(results) == 0 {
		return nil, errorz.NotFound
	}

	promo = results[0].Promo

	// Уникальные категории и промокоды
	categoryMap := make(map[string]entity.Category)
	promoUniqueMap := make(map[string]entity.PromoUnique)

	for _, r := range results {
		if r.CategoryID != "" {
			categoryMap[r.CategoryID] = entity.Category{
				CategoryID: r.CategoryID,
				PromoID:    id,
				Name:       r.Name,
			}
		}

		if r.PromoUniqueID != "" {
			promoUniqueMap[r.PromoUniqueID] = entity.PromoUnique{
				PromoUniqueID: r.PromoUniqueID,
				PromoID:       id,
				Body:          r.Body,
				Activated:     r.Activated,
			}
		}
	}

	// Конвертация мап в массивы
	for _, category := range categoryMap {
		promo.Categories = append(promo.Categories, category)
	}

	for _, promoUnique := range promoUniqueMap {
		promo.PromoUnique = append(promo.PromoUnique, promoUnique)
	}

	return &promo, nil
}

// GetAll is a method that returns a slice of pointers to all Promo instances.
func (s *promoStorage) GetAll(ctx context.Context, limit, offset int) ([]entity.Promo, error) {
	var promos []entity.Promo
	err := s.db.WithContext(ctx).Model(&entity.Promo{}).Limit(limit).Offset(offset).Find(&promos).Error
	return promos, err
}

func (s *promoStorage) GetWithPagination(ctx context.Context, limit, offset int, sortBy string, countries []countries.CountryCode) ([]entity.Promo, int64, error) {
	var promos []entity.Promo
	var total int64
	query := s.db.WithContext(ctx).Model(&entity.Promo{}).Limit(limit).Offset(offset).Order(sortBy).Find(&promos)
	err := query.Error
	s.db.WithContext(ctx).Model(&entity.Promo{}).Count(&total)
	if len(countries) == 0 {
		return promos, total, err
	}
	for i, promo := range promos {
		if !slices.Contains(countries, promo.Country) {
			promos = append(promos[:i], promos[i+1:]...)
		}
	}
	return promos, total, err
}

// Update is a method to update an existing Promo in database.
func (s *promoStorage) Update(ctx context.Context, fiberCtx fiber.Ctx, promo *entity.Promo) (*entity.Promo, error) {
	var oldPromo entity.Promo
	s.db.WithContext(ctx).Model(&entity.Promo{}).Where("id = ?", promo.PromoID).First(&oldPromo)
	if oldPromo.CompanyID != fiberCtx.Locals("business").(*entity.Business).ID {
		return nil, errorz.Forbidden
	}
	query := s.db.WithContext(ctx).Model(&entity.Promo{}).Where("id = ?", promo.PromoID).Updates(&promo)
	if query.RowsAffected == 0 {
		return nil, errorz.NotFound
	}
	return promo, query.Error
}

func (s *promoStorage) getFeed() {

}
