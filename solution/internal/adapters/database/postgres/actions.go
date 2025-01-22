package postgres

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type actionsStorage struct {
	db *gorm.DB
}

func NewActionsStorage(db *gorm.DB) *actionsStorage {
	return &actionsStorage{db: db}
}

func (s *actionsStorage) IsLikedByUser(ctx context.Context, userID, promoID string) bool {
	query := `
		SELECT COUNT(*)
		FROM actions a
				 INNER JOIN promos p ON p.promo_id = a.promo_id
				 INNER JOIN users u ON u.id = a.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?
		  AND a."like" = true`

	var total int64
	s.db.WithContext(ctx).Raw(query, userID, promoID).Scan(&total)

	return total != 0
}

func (s *actionsStorage) AddLike(ctx context.Context, userID, promoID string) error {
	querySelect := `
		SELECT COUNT(*)
		FROM actions a
				 INNER JOIN promos p ON p.promo_id = a.promo_id
				 INNER JOIN users u ON u.id = a.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?`

	queryInsert := `INSERT INTO actions (user_id, promo_id, "like") VALUES (?, ?, true)`

	queryUpdate := `UPDATE actions SET "like" = true WHERE user_id = ? AND promo_id = ?`

	queryIncrement := `UPDATE promos SET like_count = like_count + 1 WHERE promo_id = ?`

	var total int64

	// if actions record doesn't exist
	if _ = s.db.WithContext(ctx).Exec(querySelect, userID, promoID).Scan(&total); total == 0 {
		err := s.db.WithContext(ctx).Exec(queryInsert, userID, promoID).Error
		if err != nil {
			return err
		}
	} else {
		err := s.db.WithContext(ctx).Exec(queryUpdate, userID, promoID).Error
		if err != nil {
			return err
		}
	}

	if err := s.db.WithContext(ctx).Exec(queryIncrement, promoID).Error; err != nil {
		return err
	}

	return nil
}

func (s *actionsStorage) DeleteLike(ctx context.Context, userID, promoID string) error {
	querySelect := `
		SELECT COUNT(*)
		FROM actions a
				 INNER JOIN promos p ON p.promo_id = a.promo_id
				 INNER JOIN users u ON u.id = a.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?
		  AND a."like" = true?`

	queryInsert := `INSERT INTO actions (user_id, promo_id, "like") VALUES (?, ?, false)`

	queryUpdate := `UPDATE actions SET "like" = false WHERE user_id = ? AND promo_id = ?`

	queryDecrement := `UPDATE promos SET like_count = like_count - 1 WHERE promo_id = ?`

	// if actions record doesn't exist
	if rowsAffected := s.db.WithContext(ctx).Exec(querySelect, userID, promoID).RowsAffected; rowsAffected == 0 {
		err := s.db.WithContext(ctx).Exec(queryInsert, userID, promoID).Error
		if err != nil {
			return err
		}
	} else {
		err := s.db.WithContext(ctx).Exec(queryUpdate, userID, promoID).Error
		if err != nil {
			return err
		}
		err = s.db.WithContext(ctx).Exec(queryDecrement, promoID).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *actionsStorage) AddComment(ctx context.Context, userID, promoID, text string) error {
	query := `INSERT INTO comments (created_at, promo_id, user_id, text) VALUES (?, ?, ?, ?)`

	queryIncrement := `UPDATE promos SET comment_count = comment_count + 1 WHERE promo_id = ?`

	err := s.db.WithContext(ctx).Exec(query, time.Now(), promoID, userID, text).Error
	if err != nil {
		return err
	}

	err = s.db.WithContext(ctx).Exec(queryIncrement, promoID).Error
	if err != nil {
		return err
	}

	return nil
}
