package postgres

import (
	"context"
	"encoding/json"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"time"
)

// promoStorage is a struct that contains a pointer to a gorm.DB instance to interact with promo repository.
type promoStorage struct {
	db             *gorm.DB
	actionsStorage *actionsStorage
}

// NewPromoStorage is a function that returns a new instance of promoStorage.
func NewPromoStorage(db *gorm.DB) *promoStorage {
	return &promoStorage{
		db:             db,
		actionsStorage: NewActionsStorage(db),
	}
}

// Create is a method to create a new Promo in database.
// TODO country with given register
func (s *promoStorage) Create(ctx context.Context, promo entity.Promo) (*entity.Promo, error) {
	// Insert a promo (parent)'s entity
	insertPromoQuery := s.db.WithContext(ctx).Raw(
		"INSERT INTO promos (company_id, created_at, updated_at, active_from, active_until, description, image_url, max_count, mode, promo_common, age_from, age_until, country) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING promo_id;",
		promo.CompanyID, time.Now(), promo.UpdatedAt, promo.ActiveFrom, promo.ActiveUntil, promo.Description, promo.ImageURL, promo.MaxCount, promo.Mode, promo.PromoCommon, promo.AgeFrom, promo.AgeUntil, promo.Country).Scan(&promo.PromoID)
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
	if err := s.db.WithContext(ctx).Raw(query, id).Scan(&results).Error; err != nil {
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

func (s *promoStorage) GetWithPagination(ctx context.Context, limit, offset int, sortBy, companyId string, countriesSlice []countries.CountryCode) ([]entity.Promo, int64, error) {
	query := `
		SELECT
			p.promo_id,
			p.company_id,
			p.created_at,
			p.updated_at,
			p.active,
			p.active_from,
			p.active_until,
			p.description,
			p.image_url,
			p.max_count,
			p.mode,
			p.like_count,
			p.used_count,
			p.promo_common,
			p.age_from,
			p.age_until,
			p.country,
			JSON_AGG(
					DISTINCT jsonb_build_object(
					'category_id', c.category_id,
					'category_name', c.name
							 )
			) AS categories,
			JSON_AGG(
					DISTINCT jsonb_build_object(
					'promo_unique_id', pu.promo_unique_id,
					'promo_unique_body', pu.body,
					'promo_unique_activated', pu.activated
							 )
			) AS promo_uniques
		FROM promos p
				 LEFT JOIN categories c ON p.promo_id = c.promo_id
				 LEFT JOIN promo_uniques pu ON p.promo_id = pu.promo_id
		WHERE
			p.company_id = ?`

	if len(countriesSlice) > 0 {
		query += ` AND (p.country IN ? OR p.country = 0)`
	}

	query += `
		GROUP BY
			p.promo_id,
			p.company_id,
			p.created_at,
			p.updated_at,
			p.active,
			p.active_from,
			p.active_until,
			p.description,
			p.image_url,
			p.max_count,
			p.mode,
			p.like_count,
			p.used_count,
			p.promo_common,
			p.age_from,
			p.age_until,
			p.country`

	switch sortBy {
	case "active_from":
		query += ` ORDER BY p.active_from DESC`
	case "active_until":
		query += ` ORDER BY p.active_until DESC`
	default:
		query += ` ORDER BY p.created_at DESC`
	}

	query += ` LIMIT ? OFFSET ?`

	type result struct {
		PromoID      string
		CompanyID    string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		Active       bool
		ActiveFrom   time.Time
		ActiveUntil  time.Time
		Description  string
		ImageURL     string
		MaxCount     int
		Mode         string
		LikeCount    int
		UsedCount    int
		PromoCommon  string
		AgeFrom      int
		AgeUntil     int
		Country      countries.CountryCode
		Categories   *string
		PromoUniques *string
	}

	type Category struct {
		CategoryID   string `json:"category_id"`
		CategoryName string `json:"category_name"`
	}

	type PromoUnique struct {
		PromoUniqueID        string `json:"promo_unique_id"`
		PromoUniqueBody      string `json:"promo_unique_body"`
		PromoUniqueActivated bool   `json:"promo_unique_activated"`
	}

	var results []result

	if len(countriesSlice) > 0 {
		if err := s.db.WithContext(ctx).Raw(query, companyId, countriesSlice, limit, offset).Scan(&results).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := s.db.WithContext(ctx).Raw(query, companyId, limit, offset).Scan(&results).Error; err != nil {
			return nil, 0, err
		}
	}

	var promos []entity.Promo

	for _, r := range results {
		var categories []Category
		var promoUniques []PromoUnique

		if r.Categories != nil {
			if err := json.Unmarshal([]byte(*r.Categories), &categories); err != nil {
				return nil, 0, err
			}
		}

		if r.PromoUniques != nil {
			if err := json.Unmarshal([]byte(*r.PromoUniques), &promoUniques); err != nil {
				return nil, 0, err
			}
		}

		promo := entity.Promo{
			PromoID:     r.PromoID,
			CompanyID:   r.CompanyID,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			Active:      r.Active,
			ActiveFrom:  r.ActiveFrom,
			ActiveUntil: r.ActiveUntil,
			Description: r.Description,
			ImageURL:    r.ImageURL,
			MaxCount:    r.MaxCount,
			Mode:        r.Mode,
			LikeCount:   r.LikeCount,
			UsedCount:   r.UsedCount,
			PromoCommon: r.PromoCommon,
			AgeFrom:     r.AgeFrom,
			AgeUntil:    r.AgeUntil,
			Country:     r.Country,
		}

		for _, category := range categories {
			promo.Categories = append(promo.Categories, entity.Category{
				CategoryID: category.CategoryID,
				Name:       category.CategoryName,
			})
		}

		for _, promoUnique := range promoUniques {
			promo.PromoUnique = append(promo.PromoUnique, entity.PromoUnique{
				PromoUniqueID: promoUnique.PromoUniqueID,
				Body:          promoUnique.PromoUniqueBody,
				Activated:     promoUnique.PromoUniqueActivated,
			})
		}

		promos = append(promos, promo)
	}

	// Получаем общее количество записей
	var total int64
	if err := s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM promos WHERE company_id = ? AND country IN ? OR country = null", companyId, countriesSlice).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	return promos, total, nil
}

// Update is a method to update an existing Promo in database.
func (s *promoStorage) Update(ctx context.Context, fiberCtx fiber.Ctx, promo *entity.Promo, id string) (*entity.Promo, error) {
	var oldPromo entity.Promo

	var total int64
	if _ = s.db.WithContext(ctx).Model(&entity.Promo{}).Where("promo_id = ?", id).Count(&total); total == 0 {
		return nil, errorz.NotFound
	}

	s.db.WithContext(ctx).Model(&entity.Promo{}).Where("promo_id = ?", id).First(&oldPromo)

	if oldPromo.CompanyID != fiberCtx.Locals("business").(*entity.Business).ID {
		return nil, errorz.Forbidden
	}

	query := s.db.WithContext(ctx).Model(&entity.Promo{}).Where("promo_id = ?", promo.PromoID).Updates(&promo)

	if query.RowsAffected == 0 {
		return nil, errorz.NotFound
	}

	return promo, query.Error
}

func (s *promoStorage) GetFeed(ctx context.Context, age, limit, offset int, country countries.CountryCode, category, active, userID string) ([]dto.PromoForUser, int64, error) {
	query := `
		SELECT p.promo_id,
			   p.company_id,
			   p.description,
			   p.image_url,
			   p.active,
			   p.like_count,
			   p.comment_count,
			   EXISTS(SELECT * from activations a WHERE a.user_id = ? AND a.promo_id = p.promo_id) AS is_activated,
			   c.name AS category_name,
			   b.name AS business_name,
			   b.id   AS business_id
		FROM promos p
				 INNER JOIN categories c ON c.name = ? AND c.promo_id = p.promo_id
				 INNER JOIN businesses b ON b.id = p.company_id
		WHERE p.age_from <= ?
		  AND p.age_until >= ?
		  AND p.country = ?
		LIMIT ? OFFSET ?`

	queryCount := `
		SELECT COUNT(*)
		FROM promos p
				 INNER JOIN categories c ON c.name = ? AND c.promo_id = p.promo_id
				 INNER JOIN businesses b ON b.id = p.company_id
		WHERE p.age_from <= ?
		  AND p.age_until >= ?
		  AND p.country = ?`

	if active != "" {
		query = `
			SELECT p.promo_id,
				   p.company_id,
				   p.description,
				   p.image_url,
				   p.active,
				   p.like_count,
				   p.comment_count,
				   EXISTS(SELECT * from activations a WHERE a.user_id = ? AND a.promo_id = p.promo_id) AS is_activated,
				   c.name AS category_name,
				   b.name AS business_name,
				   b.id   AS business_id
			FROM promos p
					 INNER JOIN categories c ON c.name = ? AND c.promo_id = p.promo_id
					 INNER JOIN businesses b ON b.id = p.company_id
			WHERE p.age_from <= ?
			  AND p.age_until >= ?
			  AND p.country = ?
			  AND p.active = ?
			LIMIT ? OFFSET ?`

		queryCount = `
		SELECT COUNT(*)
		FROM promos p
				 INNER JOIN categories c ON c.name = ? AND c.promo_id = p.promo_id
				 INNER JOIN businesses b ON b.id = p.company_id
		WHERE p.age_from <= ?
		  AND p.age_until >= ?
		  AND p.country = ?
          AND p.active = ?`
	}

	type result struct {
		PromoID      string
		BusinessID   string
		BusinessName string
		Description  string
		ImageURL     string
		Active       bool
		LikeCount    int
		CommentCount int
		IsActivated  bool
		CategoryName string
	}

	var results []result

	if active != "" {
		if err := s.db.WithContext(ctx).Raw(query, userID, category, age, age, country, active, limit, offset).Scan(&results).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := s.db.WithContext(ctx).Raw(query, userID, category, age, age, country, limit, offset).Scan(&results).Error; err != nil {
			return nil, 0, err
		}
	}

	var promos []dto.PromoForUser

	for _, r := range results {
		promos = append(promos, dto.PromoForUser{
			PromoID:           r.PromoID,
			CompanyID:         r.BusinessID,
			CompanyName:       r.BusinessName,
			Description:       r.Description,
			ImageURL:          r.ImageURL,
			Active:            r.Active,
			IsLikedByUser:     s.actionsStorage.IsLikedByUser(ctx, userID, r.PromoID),
			IsActivatedByUser: r.IsActivated,
			LikeCount:         r.LikeCount,
		})
	}

	var total int64
	if active != "" {
		if err := s.db.WithContext(ctx).Raw(queryCount, category, age, age, country, active).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := s.db.WithContext(ctx).Raw(queryCount, category, age, age, country).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	return promos, total, nil
}

// GetByIdUser Get promo by ID for Users
func (s *promoStorage) GetByIdUser(ctx context.Context, promoID, userID string) (dto.PromoForUser, error) {
	var promo dto.PromoForUser

	query := `
		SELECT p.promo_id,
			   p.company_id,
			   p.description,
			   p.image_url,
			   p.active,
			   p.like_count,
			   p.comment_count,
			   EXISTS(SELECT * from activations a WHERE a.user_id = ? AND a.promo_id = ?) AS is_activated, -- is_activated_by_user
			   b.name AS business_name,
			   b.id   AS business_id
		FROM promos p
				 INNER JOIN businesses b ON b.id = p.company_id
		WHERE p.promo_id = ?;`

	type result struct {
		PromoID      string
		BusinessID   string
		BusinessName string
		Description  string
		ImageURL     string
		Active       bool
		LikeCount    int
		CommentCount int
		IsActivated  bool
	}

	var queryResult result

	if err := s.db.WithContext(ctx).Raw(query, userID, promoID, promoID).Scan(&queryResult).Error; err != nil {
		return promo, err
	}

	promo = dto.PromoForUser{
		PromoID:           queryResult.PromoID,
		CompanyID:         queryResult.BusinessID,
		CompanyName:       queryResult.BusinessName,
		Description:       queryResult.Description,
		ImageURL:          queryResult.ImageURL,
		Active:            queryResult.Active,
		LikeCount:         queryResult.LikeCount,
		IsLikedByUser:     s.actionsStorage.IsLikedByUser(ctx, userID, promoID),
		IsActivatedByUser: queryResult.IsActivated,
		UsedCount:         queryResult.CommentCount,
	}

	return promo, nil
}

func (s *promoStorage) GetHistory(ctx context.Context, userID string, limit, offset int) ([]dto.PromoForUser, int64, error) {
	query := `
		SELECT p.promo_id,
			   p.company_id,
			   p.description,
			   p.image_url,
			   p.active,
			   p.like_count,
			   p.comment_count,
			   EXISTS(SELECT * from activations a WHERE a.user_id = ? AND a.promo_id = p.promo_id) AS is_activated, -- is_activated_by_user
			   b.name                                                                     AS business_name,
			   b.id                                                                       AS business_id,
			   COUNT(*) OVER() AS total_count
		FROM activations a
				 INNER JOIN promos p on p.promo_id = a.promo_id
				 INNER JOIN businesses b ON b.id = p.company_id
		WHERE user_id = ?
		LIMIT ? OFFSET ?`

	type result struct {
		PromoID      string
		BusinessID   string
		BusinessName string
		Description  string
		ImageURL     string
		Active       bool
		LikeCount    int
		CommentCount int
		IsActivated  bool
		TotalCount   int64
	}

	var results []result
	if err := s.db.WithContext(ctx).Raw(query, userID, userID, limit, offset).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	var promos []dto.PromoForUser
	for _, r := range results {
		promos = append(promos, dto.PromoForUser{
			PromoID:           r.PromoID,
			CompanyID:         r.BusinessID,
			CompanyName:       r.BusinessName,
			Description:       r.Description,
			ImageURL:          r.ImageURL,
			Active:            r.Active,
			LikeCount:         r.LikeCount,
			IsLikedByUser:     s.actionsStorage.IsLikedByUser(ctx, userID, r.PromoID),
			IsActivatedByUser: r.IsActivated,
			UsedCount:         r.CommentCount,
		})
	}

	return promos, results[0].TotalCount, nil
}

func (s *promoStorage) GetStats(ctx context.Context, promoID, companyID string) (dto.PromoStatsResponse, error) {
	query := `
		SELECT DISTINCT p.used_count,
			   COUNT(*) OVER (PARTITION BY u.country) AS activations_count,
			   u.country
		FROM promos p
				 INNER JOIN public.activations a on p.promo_id = a.promo_id
				 INNER JOIN users u ON a.user_id = u.id
		WHERE company_id = ?
		AND p.promo_id = ?`

	type result struct {
		ActivationsCount int
		UsedCount        int
		Country          int
	}

	var results []result
	if err := s.db.WithContext(ctx).Raw(query, companyID, promoID).Scan(&results).Error; err != nil {
		return dto.PromoStatsResponse{}, err
	}

	if len(results) == 0 {
		return dto.PromoStatsResponse{}, errorz.NotFound
	}

	var stats dto.PromoStatsResponse
	for _, r := range results {
		stats.ActivationsCount = results[0].ActivationsCount
		stats.Countries = append(stats.Countries, dto.ActivationsByCountry{
			Country: countries.CountryCode(r.Country).Alpha2(),
			Count:   r.ActivationsCount,
		})
	}

	return stats, nil
}
