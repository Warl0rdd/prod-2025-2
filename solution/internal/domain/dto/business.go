package dto

type BusinessRegister struct {
	Email    string `json:"email" validate:"required,email,min=8,max=120" example:"example@gmail.com"` // Required, email must be valid
	Password string `json:"password" validate:"required,password,min=8,max=60" example:"Password1234"` // Required, password must meet certain requirements: must has upper case letters, lower case letters and digits
	Name     string `json:"name" validate:"required,min=5,max=50" example:"linuxflight"`               // Required, user's username
}

type BusinessRegisterResponse struct {
	BusinessID string `json:"company_id"` // User object
	Token      string `json:"token"`      // Access token
}

type BusinessLogin struct {
	Email    string `json:"email" validate:"required,email,min=8,max=120" example:"example@gmail.com"` // User's email, must be valid email address
	Password string `json:"password" validate:"required,password,min=8,max=60" example:"Password1234"` // User's password
}

type BusinessLoginResponse struct {
	Token string `json:"token"` // Access token
}
