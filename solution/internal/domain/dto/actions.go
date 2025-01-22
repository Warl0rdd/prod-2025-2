package dto

type AddLike struct {
	PromoID string `uri:"id" validate:"required"`
}

type AddComment struct {
	PromoID string `uri:"id" validate:"required"`
	Text    string `json:"text" validate:"required"`
}
