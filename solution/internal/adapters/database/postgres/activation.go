package postgres

import (
	"context"
	"github.com/biter777/countries"
	"gorm.io/gorm"
	"solution/internal/domain/common/errorz"
)

type activationStorage struct {
	db *gorm.DB
}

func NewActivationStorage(db *gorm.DB) *activationStorage {
	return &activationStorage{db: db}
}

func (s *activationStorage) ActivatePromo(ctx context.Context, age int, country countries.CountryCode, promoID, userID string) (string, error) {
	queryCount := `SELECT count(*) FROM promos WHERE promo_id = ?`

	queryActivate := `
		WITH common_update AS (
			-- Update for COMMON mode
			UPDATE promos
				SET
					used_count = CASE
									 WHEN active = true
										 AND age_from <= ?
										 AND age_until >= ?
										 AND (country = ? OR country = 0)
										 AND mode = 'COMMON'
										 AND max_count > used_count
										 AND promo_id = ? THEN used_count + 1
									 ELSE used_count
						END,
					active = CASE
								 WHEN (used_count >= max_count
										  AND mode = 'COMMON')
									 OR active_until < now() THEN false
								 ELSE true
						END
				WHERE promo_id = ?
				RETURNING promo_common AS promocode),
			 selected_unique AS (
				 -- Select UNIQUE promo code
				 SELECT pu.promo_unique_id, pu.body
				 FROM promo_uniques pu
						  INNER JOIN promos p ON p.promo_id = pu.promo_id
				 WHERE p.active = TRUE
				   AND p.age_from <= ?
				   AND p.age_until >= ?
				   AND (p.country = ? OR p.country = 0)
				   AND p.mode = 'UNIQUE'
				   AND p.promo_id = ?
				   AND pu.activated = FALSE
				 LIMIT 1 FOR UPDATE SKIP LOCKED),
			 unique_update AS (
				 -- Update UNIQUE promo code
				 UPDATE promo_uniques
					 SET activated = TRUE
					 WHERE promo_unique_id IN (SELECT promo_unique_id FROM selected_unique)
					 RETURNING body AS promocode),
			 unique_active_update AS (
				 -- Update active status for UNIQUE promos
				 UPDATE promos
					 SET active = CASE
									  WHEN EXISTS (SELECT *
												   FROM promo_uniques pu
												   WHERE pu.promo_id = promos.promo_id
													 AND pu.activated = FALSE) AND active_until > now() THEN true
									  ELSE false
						 END
					 WHERE promo_id = ?
						 AND mode = 'UNIQUE'
					 RETURNING NULL -- Ensure query consistency
			 )`

	querySelect := `SELECT mode from promos WHERE promo_id = ?`

	type selectResult struct {
		Mode string
	}

	var selectRes selectResult

	if err := s.db.WithContext(ctx).Raw(querySelect, promoID).Scan(&selectRes).Error; err != nil {
		return "", err
	}

	if selectRes.Mode == "COMMON" {
		queryActivate += `
			SELECT * from common_update
			LIMIT 1`
	} else {
		queryActivate += `SELECT * from unique_update`
	}

	type result struct {
		Promocode string
	}

	var res result
	var promosCount int64

	s.db.WithContext(ctx).Raw(queryCount, promoID).Scan(&promosCount)

	if promosCount == 0 {
		return "", errorz.NotFound
	}

	if err := s.db.Raw(queryActivate, age, age, country, promoID, promoID, age, age, country, promoID, promoID).Scan(&res).Error; err != nil {
		return "", err
	}

	if res == (result{}) {
		return "", errorz.Forbidden
	}

	err := s.db.WithContext(ctx).Exec(`UPDATE promos SET used_count = used_count + 1 WHERE promo_id = ?`, promoID).Error
	if err != nil {
		return "", err
	}

	err = s.db.WithContext(ctx).Exec(`INSERT INTO activations (user_id, promo_id) VALUES (?, ?)`, userID, promoID).Error
	if err != nil {
		return "", err
	}

	return res.Promocode, nil
}
