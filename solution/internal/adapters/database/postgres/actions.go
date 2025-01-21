package postgres

import (
	"context"
	"gorm.io/gorm"
)

type actionsStorage struct {
	db *gorm.DB
}

func NewActionsStorage(db *gorm.DB) *actionsStorage {
	return &actionsStorage{db: db}
}

func (s *actionsStorage) AddLike(ctx context.Context, userID, promoID string) error {
	querySelect := `
		SELECT *
		FROM actions a
				 INNER JOIN promos p ON p.promo_id = a.promo_id
				 INNER JOIN users u ON u.id = a.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?`

	queryInsert := `INSERT INTO actions (user_id, promo_id, "like") VALUES (?, ?, true)`

	queryUpdate := `UPDATE actions SET "like" = true WHERE user_id = ? AND promo_id = ?`

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
	}

	return nil
}
