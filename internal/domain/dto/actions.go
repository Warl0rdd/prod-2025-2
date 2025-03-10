package dto

type AddLike struct {
	PromoID string `uri:"id" validate:"required"`
}

type AddComment struct {
	PromoID string `uri:"id" validate:"required"`
	Text    string `json:"text" validate:"required,min=10,max=1000"`
}

type GetComments struct {
	ID     string `uri:"id" validate:"required"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

type Comment struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Date   string `json:"date"`
	Author Author `json:"author"`
}

type Author struct {
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type GetCommentById struct {
	ID        string `uri:"id" validate:"required"` // promo id
	CommentID string `uri:"comment_id" validate:"required"`
}

type DeleteCommentById struct {
	ID        string `uri:"id" validate:"required"` // promo id
	CommentID string `uri:"comment_id" validate:"required"`
}

type UpdateComment struct {
	ID        string `uri:"id" validate:"required"` // promo id
	CommentID string `uri:"comment_id" validate:"required"`
	Text      string `json:"text" validate:"required,min=10,max=1000"`
}

type Activate struct {
	ID string `uri:"id" validate:"required"` // promo id
}

type ActivateResponse struct {
	Promo string `json:"promo,omitempty"`
}
