package dto

type PromoCreate struct {
	CompanyID string `json:"company_id"`

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
