package entity

import (
	"github.com/biter777/countries"
	"time"
)

type Promo struct {
	PromoID   string    `json:"promo_id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"not null;foreignKey:ID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Active       bool          `json:"-" gorm:"default:true"`
	ActiveFrom   time.Time     `json:"active_from" gorm:"default:now()"`
	ActiveUntil  time.Time     `json:"active_until" gorm:"default:'2525-01-22 13:53:19.177440 +00:00'"` // очень очень далёкая дата
	Description  string        `json:"description" gorm:"not null"`
	ImageURL     string        `json:"image_url"`
	MaxCount     int           `json:"max_count" gorm:"not null"`
	Mode         string        `json:"mode" gorm:"not null"`
	LikeCount    int           `json:"like_count" gorm:"default:0"`
	UsedCount    int           `json:"used_count" gorm:"default:0"`
	CommentCount int           `json:"comment_count" gorm:"default:0"`
	PromoCommon  string        `json:"promo_common"`
	PromoUnique  []PromoUnique `json:"promo_unique;" gorm:"foreignKey:PromoID"`

	AgeFrom         int                   `json:"age_from"`
	AgeUntil        int                   `json:"age_until"`
	Country         countries.CountryCode `json:"country"`
	CountryOriginal string                `json:"-" gorm:"country_original"`
	Categories      []Category            `json:"categories" gorm:"foreignKey:PromoID"`
	Actions         []Likes               `json:"-" gorm:"foreignKey:PromoID"`
	Comments        []Comment             `json:"-" gorm:"foreignKey:PromoID"`
	Activations     []Activation          `json:"-" gorm:"foreignKey:PromoID"`
}

type PromoUnique struct {
	PromoUniqueID string `json:"-" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	PromoID       string `json:"-" gorm:"not null;"`
	Body          string `json:"-" gorm:"not null"`
	Activated     bool   `json:"-" gorm:"default:false"`
	Index         int    `json:"-"`
}

type Category struct {
	CategoryID string `json:"id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	PromoID    string `json:"-" gorm:"not null;"`
	Name       string `json:"name" gorm:"not null"`
	Index      int    `json:"-"`
}
