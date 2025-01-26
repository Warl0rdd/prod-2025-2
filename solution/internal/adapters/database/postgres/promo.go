package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"solution/internal/adapters/logger"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/utils/pointers"
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
		"INSERT INTO promos (company_id, created_at, updated_at, active_from, active_until, description, image_url, max_count, mode, promo_common, age_from, age_until, country, active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING promo_id;",
		promo.CompanyID, time.Now(), promo.UpdatedAt, promo.ActiveFrom, promo.ActiveUntil, promo.Description, promo.ImageURL, promo.MaxCount, promo.Mode, promo.PromoCommon, promo.AgeFrom, promo.AgeUntil, promo.Country, promo.Active).Scan(&promo.PromoID)
	if err := insertPromoQuery.Error; err != nil {
		return nil, err
	}

	// Insert categories
	for i, category := range promo.Categories {
		insertCategoryQuery := s.db.WithContext(ctx).Exec("INSERT INTO categories (promo_id, name, index) VALUES (?, ?, ?);", promo.PromoID, category.Name, i)
		if err := insertCategoryQuery.Error; err != nil {
			return nil, err
		}
	}

	// Insert promo_uniques
	for i, promoUnique := range promo.PromoUnique {
		insertPromoUniqueQuery := s.db.WithContext(ctx).Exec("INSERT INTO promo_uniques (promo_id, body, activated, index) VALUES (?, ?, ?, ?);", promo.PromoID, promoUnique.Body, promoUnique.Activated, i)
		if err := insertPromoUniqueQuery.Error; err != nil {
			return nil, err
		}
	}

	return &promo, nil
}

// GetByID is a method that returns an error and a pointer to a Promo instance by id.
func (s *promoStorage) GetByID(ctx context.Context, promoId string) (*entity.Promo, error) {
	query := `
        SELECT p.promo_id,
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
               COALESCE(
                           JSONB_AGG(
                           jsonb_build_object(
                                   'category_id', c.category_id,
                                   'category_name', c.name,
                                   'index', c.index
                           ) ORDER BY c.index
                                   ) FILTER (WHERE c.category_id IS NOT NULL),
                           '[]'::jsonb
               ) AS categories,
        
               COALESCE(
                           JSONB_AGG(
                           jsonb_build_object(
                                   'promo_unique_id', pu.promo_unique_id,
                                   'promo_unique_body', pu.body,
                                   'promo_unique_activated', pu.activated,
                                   'index', pu.index
                           ) ORDER BY pu.index
                                   ) FILTER (WHERE pu.promo_unique_id IS NOT NULL),
                           '[]'::jsonb
               ) AS promo_uniques
        FROM promos p
                 LEFT JOIN categories c ON p.promo_id = c.promo_id
                 LEFT JOIN promo_uniques pu ON p.promo_id = pu.promo_id
        WHERE p.promo_id = ?
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

	var res result
	if err := s.db.WithContext(ctx).Raw(query, promoId).Scan(&res).Error; err != nil {
		return nil, err
	}

	if res.PromoID == "" {
		return nil, nil
	}

	var categories []struct {
		CategoryID   string `json:"category_id"`
		CategoryName string `json:"category_name"`
		Index        int    `json:"index"`
	}

	if res.Categories != nil {
		if err := json.Unmarshal([]byte(*res.Categories), &categories); err != nil {
			return nil, err
		}
	}

	var promoUniques []struct {
		PromoUniqueID        string `json:"promo_unique_id"`
		PromoUniqueBody      string `json:"promo_unique_body"`
		PromoUniqueActivated bool   `json:"promo_unique_activated"`
		Index                int    `json:"index"`
	}

	if res.PromoUniques != nil {
		if err := json.Unmarshal([]byte(*res.PromoUniques), &promoUniques); err != nil {
			return nil, err
		}
	}

	promo := &entity.Promo{
		PromoID:     res.PromoID,
		CompanyID:   res.CompanyID,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
		Active:      res.Active,
		ActiveFrom:  res.ActiveFrom,
		ActiveUntil: res.ActiveUntil,
		Description: res.Description,
		ImageURL:    res.ImageURL,
		MaxCount:    res.MaxCount,
		Mode:        res.Mode,
		LikeCount:   res.LikeCount,
		UsedCount:   res.UsedCount,
		PromoCommon: res.PromoCommon,
		AgeFrom:     res.AgeFrom,
		AgeUntil:    res.AgeUntil,
		Country:     res.Country,
	}

	for _, category := range categories {
		promo.Categories = append(promo.Categories, entity.Category{
			CategoryID: category.CategoryID,
			Name:       category.CategoryName,
		})
	}

	for _, pu := range promoUniques {
		promo.PromoUnique = append(promo.PromoUnique, entity.PromoUnique{
			PromoUniqueID: pu.PromoUniqueID,
			Body:          pu.PromoUniqueBody,
			Activated:     pu.PromoUniqueActivated,
		})
	}

	return promo, nil
}

func (s *promoStorage) GetWithPagination(ctx context.Context, limit, offset int, sortBy, companyId string, countriesSlice []countries.CountryCode) ([]entity.Promo, int64, error) {
	query := `
		SELECT p.promo_id,
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
			   COALESCE(
							   JSONB_AGG(
							   jsonb_build_object(
									   'category_id', c.category_id,
									   'category_name', c.name,
									   'index', c.index
							   ) ORDER BY c.index
									   ) FILTER (WHERE c.category_id IS NOT NULL),
							   '[]'::jsonb
			   ) AS categories,
		
			   -- Уникальные промо с сортировкой по index
			   COALESCE(
							   JSONB_AGG(
							   jsonb_build_object(
									   'promo_unique_id', pu.promo_unique_id,
									   'promo_unique_body', pu.body,
									   'promo_unique_activated', pu.activated,
									   'index', pu.index
							   ) ORDER BY pu.index
									   ) FILTER (WHERE pu.promo_unique_id IS NOT NULL),
							   '[]'::jsonb
			   ) AS promo_uniques
		FROM promos p
				 LEFT JOIN categories c ON p.promo_id = c.promo_id
				 LEFT JOIN promo_uniques pu ON p.promo_id = pu.promo_id
		WHERE p.company_id = ?`

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
		Index        int    `json:"index"`
	}

	type PromoUnique struct {
		PromoUniqueID        string `json:"promo_unique_id"`
		PromoUniqueBody      string `json:"promo_unique_body"`
		PromoUniqueActivated bool   `json:"promo_unique_activated"`
		Index                int    `json:"index"`
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
	if len(countriesSlice) > 0 {
		if err := s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM promos WHERE company_id = ? AND (country IN ? OR country = 0)", companyId, countriesSlice).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := s.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM promos WHERE company_id = ?", companyId).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	return promos, total, nil
}

// Update is a method to update an existing Promo in database.
func (s *promoStorage) Update(ctx context.Context, fiberCtx fiber.Ctx, promo dto.PromoUpdate, id string) (*entity.Promo, error) {
	var oldPromo entity.Promo
	companyID := fiberCtx.Locals("business").(*entity.Business).ID

	queryUpdate := `
		UPDATE promos
		SET
		    active_from = COALESCE(?, active_from),
    		active_until = COALESCE(?, active_until),
			description = COALESCE(?, description),
			image_url = COALESCE(?, image_url),
			max_count = COALESCE(?, max_count),
			mode = COALESCE(?, mode),
			promo_common = COALESCE(?, promo_common),
			age_from = COALESCE(?, age_from),
			age_until = COALESCE(?, age_until),
			active = COALESCE(?, active)`

	if promo.Target != nil && promo.Target.Country != "" {
		queryUpdate += `, country = COALESCE(?, country)`
	}

	queryUpdate += `WHERE promo_id = ?`

	queryUpdateCategories := `
		INSERT INTO categories (promo_id, name, index)
		VALUES
			(?, ?, ?)`

	queryUpdatePromoUniques := `
		INSERT INTO promo_uniques (promo_id, body, activated, index)
		VALUES
			(?, ?, ?, ?)`

	findOldPromoQuery := s.db.WithContext(ctx).Where("promo_id = ?", id).First(&oldPromo)

	if findOldPromoQuery.Error != nil {
		return nil, errorz.NotFound
	}

	if oldPromo.CompanyID != companyID {
		return nil, errorz.Forbidden
	}

	if (promo.MaxCount != nil) && oldPromo.Mode == "UNIQUE" && (promo.PromoUnique == nil || (*promo.MaxCount != 1)) {
		logger.Log.Error("unique max count")
		return nil, errorz.BadRequest
	}

	var active *bool

	currentActive := oldPromo.Active
	active = &currentActive

	if oldPromo.ActiveFrom.After(time.Now()) || oldPromo.ActiveUntil.Before(time.Now()) {
		*active = false
	}

	var activeFrom, activeUntil *time.Time
	var timeError error
	if promo.ActiveFrom != nil {
		activeFrom, timeError = pointers.Time(time.Parse("2006-01-02", *promo.ActiveFrom))
		if timeError != nil {
			//activeFrom, timeError = pointers.Time(time.Parse("2006-01-02 15:04:05", *promo.ActiveFrom))
			//if timeError != nil {
			//	return nil, timeError
			//}

			return nil, errorz.BadRequest
		}
	}
	if promo.ActiveUntil != nil {
		activeUntil, timeError = pointers.Time(time.Parse("2006-01-02", *promo.ActiveUntil))
		if timeError != nil {
			//activeUntil, timeError = pointers.Time(time.Parse("2006-01-02 15:04:05", *promo.ActiveUntil))
			//if timeError != nil {
			//	return nil, timeError
			//}

			return nil, errorz.BadRequest
		}
	}

	if activeFrom != nil && activeFrom.After(time.Now()) {
		*active = false
	}

	if activeUntil != nil && activeUntil.Before(time.Now()) {
		*active = false
	}

	var ageFrom, ageUntil *int

	if promo.Target != nil {
		if promo.Target.AgeFrom == 0 {
			ageFrom = nil
		} else {
			ageFrom = &promo.Target.AgeFrom
		}

		if promo.Target.AgeUntil == 0 {
			ageUntil = nil
		} else {
			ageUntil = &promo.Target.AgeUntil
		}
	}

	if promo.Target != nil && promo.Target.Country != "" {
		if err := s.db.WithContext(ctx).Exec(queryUpdate,
			activeFrom,
			activeUntil,
			promo.Description,
			promo.ImageURL,
			promo.MaxCount,
			promo.Mode,
			promo.PromoCommon,
			ageFrom,
			ageUntil,
			active,
			countries.ByName(promo.Target.Country),
			id).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.WithContext(ctx).Exec(queryUpdate,
			activeFrom,
			activeUntil,
			promo.Description,
			promo.ImageURL,
			promo.MaxCount,
			promo.Mode,
			promo.PromoCommon,
			ageFrom,
			ageUntil,
			active,
			id).Error; err != nil {
			return nil, err
		}
	}

	if promo.Target != nil && promo.Target.Categories != nil {
		s.db.WithContext(ctx).Exec(`DELETE FROM categories WHERE promo_id = ?`, id)
		for i, category := range promo.Target.Categories {
			if err := s.db.WithContext(ctx).Exec(queryUpdateCategories, id, category, i).Error; err != nil {
				return nil, err
			}
		}
	}

	if promo.PromoUnique != nil {
		s.db.WithContext(ctx).Exec(`DELETE FROM promo_uniques WHERE promo_id = ?`, id)
		for i, promoUnique := range promo.PromoUnique {
			if err := s.db.WithContext(ctx).Exec(queryUpdatePromoUniques, id, promoUnique, i).Error; err != nil {
				return nil, err
			}
		}
	}

	newPromo, err := s.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return newPromo, nil
}

func (s *promoStorage) GetFeed(ctx context.Context, age, limit, offset int, country countries.CountryCode, category *string, active, userID string) ([]dto.PromoForUser, int64, error) {
	baseQuery := `
        SELECT 
            p.promo_id,
            p.company_id,
            p.description,
            p.image_url,
            p.active,
            p.like_count,
            p.comment_count,
            EXISTS(SELECT 1 FROM activations a WHERE a.user_id = ? AND a.promo_id = p.promo_id) AS is_activated,
            c.name AS category_name,
            b.name AS business_name,
            b.id AS business_id
        FROM promos p
        %s JOIN categories c ON c.promo_id = p.promo_id %s
        INNER JOIN businesses b ON b.id = p.company_id
        WHERE p.age_from <= ?
          AND p.age_until >= ?
          AND (p.country = ? OR p.country = 0)
          %s  -- Условие active
        ORDER BY p.created_at DESC
        LIMIT ? OFFSET ?`

	baseCountQuery := `
        SELECT COUNT(*)
        FROM promos p
        %s JOIN categories c ON c.promo_id = p.promo_id %s
        INNER JOIN businesses b ON b.id = p.company_id
        WHERE p.age_from <= ?
          AND p.age_until >= ?
          AND (p.country = ? OR p.country = 0)
          %s --Active`

	// Определяем тип JOIN и условия для категории
	joinType := "LEFT"
	categoryCondition := ""
	if category != nil {
		joinType = "INNER"
		categoryCondition = "AND LOWER(c.name) = LOWER(?)"
	}

	// Добавляем условие active, если нужно
	activeCondition := ""
	if active != "" {
		activeCondition = "AND p.active = ?"
	}

	// Формируем итоговые запросы
	query := fmt.Sprintf(baseQuery, joinType, categoryCondition, activeCondition)
	queryCount := fmt.Sprintf(baseCountQuery, joinType, categoryCondition, activeCondition)

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
	var args []interface{}

	// Параметры запроса
	args = append(args, userID)
	if category != nil {
		args = append(args, *category)
	}
	args = append(args, age, age, country)
	if active != "" {
		args = append(args, active)
	}
	args = append(args, limit, offset)

	// Выполнение основного запроса
	if err := s.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	// Преобразование результатов
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

	var countArgs []interface{}
	if category != nil {
		countArgs = append(countArgs, *category)
	}
	countArgs = append(countArgs, age, age, country)
	if active != "" {
		countArgs = append(countArgs, active)
	}

	var total int64
	if err := s.db.WithContext(ctx).Raw(queryCount, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
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

	if promo.PromoID == "" {
		return promo, errorz.NotFound
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
