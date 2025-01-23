package dto

type AntiFraudRequest struct {
	UserEmail string `json:"user_email"`
	PromoID   string `json:"promo_id"`
}

type AntiFraudResponse struct {
	Ok         bool   `json:"ok"`
	CacheUntil string `json:"cache_until"`
}
