package dto

import (
	"github.com/biter777/countries"
	"time"
)

type PromoCreate struct {
	Target      *Target  `json:"target" validate:"required"`
	ActiveFrom  string   `json:"active_from"`
	ActiveUntil string   `json:"active_until"`
	Description string   `json:"description" validate:"required"`
	ImageURL    string   `json:"image_url" validate:"omitempty,url"`
	MaxCount    int      `json:"max_count" validate:"omitempty,required"`
	Mode        string   `json:"mode"`
	PromoCommon string   `json:"promo_common"`
	PromoUnique []string `json:"promo_unique"`
	Active      bool
}

type Target struct {
	AgeFrom    int      `json:"age_from,omitempty"`
	AgeUntil   int      `json:"age_until,omitempty"`
	Country    string   `json:"country,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

type PromoCreateResponse struct {
	ID string `json:"id"`
}

type PromoUpdate struct {
	Target      *Target  `json:"target,omitempty"`
	ActiveFrom  *string  `json:"active_from,omitempty"`
	ActiveUntil *string  `json:"active_until,omitempty"`
	Description *string  `json:"description,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty" validate:"omitempty,url"`
	MaxCount    *int     `json:"max_count,omitempty"`
	Mode        *string  `json:"mode,omitempty"`
	PromoCommon *string  `json:"promo_common,omitempty"`
	PromoUnique []string `json:"promo_unique,omitempty"`
}

type PromoGetWithPaginationRequest struct {
	Limit     int      `query:"limit"`
	Offset    int      `query:"offset"`
	SortBy    string   `query:"sort_by"`
	Countries []string `query:"country"`
}

type PromoGetWithPagination struct {
	Limit     int
	Offset    int
	SortBy    string
	Countries []countries.CountryCode
}

type PromoGetByID struct {
	ID string `uri:"id" validate:"required"`
}

type PromoDTO struct {
	PromoID     string    `json:"promo_id"`
	CompanyID   string    `json:"company_id"`
	CompanyName string    `json:"company_name"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`

	Target      Target   `json:"target"`
	Active      bool     `json:"active"`
	ActiveFrom  string   `json:"active_from"`
	ActiveUntil string   `json:"active_until"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url"`
	MaxCount    int      `json:"max_count"`
	Mode        string   `json:"mode"`
	LikeCount   int      `json:"like_count"`
	UsedCount   int      `json:"used_count"`
	PromoCommon string   `json:"promo_common"`
	PromoUnique []string `json:"promo_unique"`
}

type PromoGetWithPaginationResponse struct {
	Promos []PromoCreate
}

type PromoFeedRequest struct {
	Limit    int     `query:"limit"`
	Offset   int     `query:"offset"`
	Category *string `query:"category"`
	Active   string  `query:"active"`
}

// PromoForUser promoDTO for user's feed
type PromoForUser struct {
	PromoID           string `json:"promo_id"`
	CompanyID         string `json:"company_id"`
	CompanyName       string `json:"company_name"`
	Description       string `json:"description"`
	ImageURL          string `json:"image_url"`
	Active            bool   `json:"active"`
	LikeCount         int    `json:"like_count"`
	IsLikedByUser     bool   `json:"is_liked_by_user"`
	IsActivatedByUser bool   `json:"is_activated_by_user"`
	UsedCount         int    `json:"used_count"`
}

type PromoHistory struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

type PromoStats struct {
	Id string `uri:"id" validate:"required"`
}

type PromoStatsResponse struct {
	ActivationsCount int                    `json:"activations_count"`
	Countries        []ActivationsByCountry `json:"countries"`
}

type ActivationsByCountry struct {
	Country string `json:"country"`
	Count   int    `json:"activations_count"`
}
