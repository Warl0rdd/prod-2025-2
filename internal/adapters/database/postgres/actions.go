package postgres

import (
	"context"
	"gorm.io/gorm"
	"prod/internal/domain/common/errorz"
	"prod/internal/domain/dto"
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
		FROM likes l
				 INNER JOIN promos p ON p.promo_id = l.promo_id
				 INNER JOIN users u ON u.id = l.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?
		  AND l."like" = true`

	var total int64
	s.db.WithContext(ctx).Raw(query, userID, promoID).Scan(&total)

	return total != 0
}

func (s *actionsStorage) AddLike(ctx context.Context, userID, promoID string) error {
	querySelect := `
		SELECT COUNT(*)
		FROM likes l
				 INNER JOIN promos p ON p.promo_id = l.promo_id
				 INNER JOIN users u ON u.id = l.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?`

	queryInsert := `INSERT INTO likes (user_id, promo_id, "like") VALUES (?, ?, true)`

	queryUpdate := `
		WITH updated_likes AS (
			UPDATE likes
				SET "like" = TRUE
				WHERE user_id = ? 
          			AND promo_id = ? 
					AND "like" = FALSE
				RETURNING promo_id
		)
		UPDATE promos
		SET like_count = like_count + 1
		WHERE promo_id IN (SELECT promo_id FROM updated_likes)`

	queryIncrement := `UPDATE promos SET like_count = like_count + 1 WHERE promo_id = ?`

	var total int64

	// if actions record doesn't exist
	if _ = s.db.WithContext(ctx).Raw(querySelect, userID, promoID).Scan(&total); total == 0 {
		err := s.db.WithContext(ctx).Exec(queryInsert, userID, promoID).Error
		if err != nil {
			return err
		}
		if err = s.db.WithContext(ctx).Exec(queryIncrement, promoID).Error; err != nil {
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

func (s *actionsStorage) DeleteLike(ctx context.Context, userID, promoID string) error {
	querySelect := `
		SELECT COUNT(*)
		FROM likes l
				 INNER JOIN promos p ON p.promo_id = l.promo_id
				 INNER JOIN users u ON u.id = l.user_id
		WHERE u.id = ?
		  AND p.promo_id = ?
		  AND l."like" = true`

	queryInsert := `INSERT INTO likes (user_id, promo_id, "like") VALUES (?, ?, false)`

	queryUpdate := `UPDATE likes SET "like" = false WHERE user_id = ? AND promo_id = ?`

	queryDecrement := `UPDATE promos SET like_count = like_count - 1 WHERE promo_id = ?`

	var total int64

	// if actions record doesn't exist
	if _ = s.db.WithContext(ctx).Raw(querySelect, userID, promoID).Scan(&total); total == 0 {
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

func (s *actionsStorage) AddComment(ctx context.Context, userID, promoID, text string) (string, error) {
	query := `INSERT INTO comments (created_at, promo_id, user_id, text) VALUES (?, ?, ?, ?) RETURNING comment_id`

	queryIncrement := `UPDATE promos SET comment_count = comment_count + 1 WHERE promo_id = ?`

	var commentID string

	err := s.db.WithContext(ctx).Raw(query, time.Now(), promoID, userID, text).Scan(&commentID).Error
	if err != nil {
		return "", err
	}

	err = s.db.WithContext(ctx).Exec(queryIncrement, promoID).Error
	if err != nil {
		return "", err
	}

	return commentID, nil
}

func (s *actionsStorage) GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, int64, error) {
	query := `
		SELECT u.name,
			   u.surname,
			   u.avatar_url,
			   c.comment_id,
			   c.text,
			   c.created_at
		FROM comments c
				 INNER JOIN users u ON u.id = c.user_id
		WHERE c.promo_id = ?
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?`

	type result struct {
		Name      string
		Surname   string
		AvatarURL string
		CommentID string
		Text      string
		CreatedAt time.Time
	}

	var results []result

	err := s.db.WithContext(ctx).Raw(query, promoID, limit, offset).Scan(&results).Error

	if err != nil {
		return nil, 0, err
	}

	comments := make([]dto.Comment, 0, len(results))

	for _, r := range results {
		comments = append(comments, dto.Comment{
			ID:   r.CommentID,
			Text: r.Text,
			Date: r.CreatedAt.Format(time.RFC3339),
			Author: dto.Author{
				Name:      r.Name,
				Surname:   r.Surname,
				AvatarURL: r.AvatarURL,
			},
		})
	}

	var total int64

	if err = s.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM comments WHERE promo_id = ?`, promoID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *actionsStorage) GetCommentById(ctx context.Context, promoID, commentID string) (dto.Comment, error) {
	query := `
		SELECT u.name,
			   u.surname,
			   u.avatar_url,
			   c.comment_id,
			   c.text,
			   c.created_at
		FROM comments c
				 INNER JOIN users u ON u.id = c.user_id
		WHERE c.comment_id = ?
		  AND c.promo_id = ?`

	type result struct {
		Name      string
		Surname   string
		AvatarURL string
		CommentID string
		Text      string
		CreatedAt time.Time
	}

	var r result

	err := s.db.WithContext(ctx).Raw(query, commentID, promoID).Scan(&r).Error

	if err != nil {
		return dto.Comment{}, err
	}

	if r.CommentID == "" {
		return dto.Comment{}, errorz.NotFound
	}

	return dto.Comment{
		ID:   r.CommentID,
		Text: r.Text,
		Date: r.CreatedAt.Format(time.RFC3339),
		Author: dto.Author{
			Name:      r.Name,
			Surname:   r.Surname,
			AvatarURL: r.AvatarURL,
		},
	}, nil
}

func (s *actionsStorage) UpdateComment(ctx context.Context, promoID, commentID, userID, text string) (dto.Comment, error) {
	querySelect := `
		SELECT u.name,
			   u.surname,
			   u.avatar_url,
			   u.id,
			   c.comment_id,
			   c.text,
			   c.created_at
		FROM comments c
				 INNER JOIN users u ON u.id = c.user_id
		WHERE c.comment_id = ?
		  AND c.promo_id = ?`

	queryUpdate := `UPDATE comments SET text = ? WHERE comment_id = ? AND promo_id = ?`

	type result struct {
		Name      string
		Surname   string
		AvatarURL string
		ID        string // UserID
		CommentID string
		Text      string
		CreatedAt time.Time
	}

	var r result

	err := s.db.WithContext(ctx).Raw(querySelect, commentID, promoID).Scan(&r).Error
	if err != nil {
		return dto.Comment{}, err
	}

	if r.CommentID == "" {
		return dto.Comment{}, errorz.NotFound
	}

	if r.ID != userID {
		return dto.Comment{}, errorz.Forbidden
	}

	query := s.db.WithContext(ctx).Exec(queryUpdate, text, commentID, promoID)
	if queryErr := query.Error; queryErr != nil {
		return dto.Comment{}, err
	}

	if query.RowsAffected == 0 {
		return dto.Comment{}, errorz.NotFound
	}

	return dto.Comment{
		ID:   r.CommentID,
		Text: text,
		Date: r.CreatedAt.Format(time.RFC3339),
		Author: dto.Author{
			Name:      r.Name,
			Surname:   r.Surname,
			AvatarURL: r.AvatarURL,
		},
	}, nil
}

func (s *actionsStorage) DeleteComment(ctx context.Context, promoID, commentID, userID string) error {
	querySelect := `SELECT user_id FROM comments WHERE comment_id = ? AND promo_id = ?`

	queryDelete := `DELETE FROM comments WHERE comment_id = ? AND promo_id = ?`

	var authorID string

	var existsPromo, existsComment bool

	errExistsPromo := s.db.WithContext(ctx).Raw(`SELECT EXISTS(SELECT * FROM promos WHERE promo_id = ?)`, promoID).Scan(&existsPromo).Error

	errExistsComment := s.db.WithContext(ctx).Raw(`SELECT EXISTS(SELECT * FROM comments WHERE comment_id = ? AND promo_id = ?)`, commentID, promoID).Scan(&existsComment).Error

	if errExistsPromo != nil || errExistsComment != nil {
		return errExistsPromo
	}

	if !existsPromo || !existsComment {
		return errorz.NotFound
	}

	err := s.db.WithContext(ctx).Raw(querySelect, commentID, promoID).Scan(&authorID).Error
	if err != nil {
		return err
	}

	if authorID != userID {
		return errorz.Forbidden
	}

	query := s.db.WithContext(ctx).Exec(queryDelete, commentID, promoID)
	if queryErr := query.Error; queryErr != nil {
		return queryErr
	}

	if query.RowsAffected != 0 {
		err = s.db.WithContext(ctx).Exec(`UPDATE promos SET comment_count = comment_count - 1 WHERE promo_id = ?`, promoID).Error
		if err != nil {
			return err
		}
	}

	return nil

}
