package dto

import (
	"github.com/biter777/countries"
	"time"
)

type PromoCreate struct {
	Target      Target   `json:"target" validate:"required"`
	ActiveFrom  string   `json:"active_from"`
	ActiveUntil string   `json:"active_until"`
	Description string   `json:"description" validate:"required"`
	ImageURL    string   `json:"image_url" validate:"url"`
	MaxCount    int      `json:"max_count"`
	Mode        string   `json:"mode"`
	PromoCommon string   `json:"promo_common"`
	PromoUnique []string `json:"promo_unique"`
}

type Target struct {
	AgeFrom    int      `json:"age_from"`
	AgeUntil   int      `json:"age_until"`
	Country    string   `json:"country"`
	Categories []string `json:"categories"`
}

type PromoCreateResponse struct {
	ID string `json:"id"`
}

type PromoGetWithPaginationRequest struct {
	Limit     int      `header:"limit"`
	Offset    int      `header:"offset"`
	SortBy    string   `header:"sort_by"`
	Countries []string `header:"countries"`
}

type PromoGetWithPagination struct {
	Limit     int
	Offset    int
	SortBy    string
	Countries []countries.CountryCode
}

type PromoDTO struct {
	PromoID     string    `json:"promo_id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	CompanyID   string    `json:"company_id" gorm:"not null;foreignKey:ID"`
	CompanyName string    `json:"company_name"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`

	Target      Target    `json:"target" gorm:"foreignKey:TargetID;not null"`
	Active      bool      `json:"active" gorm:"default:true"`
	ActiveFrom  time.Time `json:"active_from"`
	ActiveUntil time.Time `json:"active_until"`
	Description string    `json:"description" gorm:"not null"`
	ImageURL    string    `json:"image_url"`
	MaxCount    int       `json:"max_count" gorm:"not null"`
	Mode        string    `json:"mode" gorm:"not null"`
	LikeCount   int       `json:"like_count" gorm:"default:0"`
	UsedCount   int       `json:"used_count" gorm:"default:0"`
	PromoCommon string    `json:"promo_common"`
	PromoUnique []string  `json:"promo_unique" gorm:"foreignKey:PromoUniqueID;"`
}

type PromoGetWithPaginationResponse struct {
	Promos []PromoCreate
}
