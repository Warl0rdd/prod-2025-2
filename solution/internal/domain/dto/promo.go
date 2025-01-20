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
	ImageURL    string   `json:"image_url" validate:"omitempty,url"`
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
	Limit     int      `query:"limit"`
	Offset    int      `query:"offset"`
	SortBy    string   `query:"sort_by"`
	Countries []string `query:"countries"`
}

type PromoGetWithPagination struct {
	Limit     int
	Offset    int
	SortBy    string
	Countries []countries.CountryCode
}

type PromoGetByID struct {
	ID string `uri:"id"`
}

type PromoDTO struct {
	PromoID     string    `json:"promo_id"`
	CompanyID   string    `json:"company_id"`
	CompanyName string    `json:"company_name"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`

	Target      Target    `json:"target"`
	Active      bool      `json:"active"`
	ActiveFrom  time.Time `json:"active_from"`
	ActiveUntil time.Time `json:"active_until"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	MaxCount    int       `json:"max_count"`
	Mode        string    `json:"mode"`
	LikeCount   int       `json:"like_count"`
	UsedCount   int       `json:"used_count"`
	PromoCommon string    `json:"promo_common"`
	PromoUnique []string  `json:"promo_unique"`
}

type PromoGetWithPaginationResponse struct {
	Promos []PromoCreate
}
