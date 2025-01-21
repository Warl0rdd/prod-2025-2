package dto

type AddLike struct {
	PromoID string `uri:"id" validate:"required"`
}
