package entity

import "time"

type Actions struct {
	ActionsID string `json:"action_id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	PromoID   string `json:"promo_id" gorm:"not null;"`
	UserID    string `json:"user_id" gorm:"not null;"`

	Like       bool `json:"like" gorm:"default:false"`
	Activation bool `json:"activation" gorm:"default:false"`
}

type Comment struct {
	CommentID string    `json:"comment_id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"date" gorm:"not null;"`
	PromoID   string    `json:"promo_id" gorm:"not null;"`
	UserID    string    `json:"user_id" gorm:"not null;"`
	Text      string    `json:"text" gorm:"not null;"`
}
