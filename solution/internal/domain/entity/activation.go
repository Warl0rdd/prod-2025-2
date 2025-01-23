package entity

type Activation struct {
	ActivationID string `json:"activation_id" gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()"`
	UserID       string `json:"user_id" gorm:"not null;"`
	PromoID      string `json:"promo_id" gorm:"not null;"`
}
