package dto

// UserRegister @Description User registration dto
type UserRegister struct {
	Email     string    `json:"email" validate:"required,email,min=8,max=120" example:"example@gmail.com"`                    // Required, email must be valid
	Password  string    `json:"password" validate:"required,password,min=8,max=60" example:"Password1234"`                    // Required, password must meet certain requirements: must have upper case letters, lower case letters and digits
	Name      string    `json:"name" validate:"required,min=1,max=100" example:"John"`                                        // Required, user's name
	Surname   string    `json:"surname" validate:"required,min=1,max=120" example:"Doe"`                                      // Required, user's surname
	AvatarURL *string   `json:"avatar_url" example:"https://example.com/avatar.jpg" validate:"omitempty,omitnil,url,max=350"` // User's avatar URL
	Other     UserOther `json:"other"`                                                                                        // User's other information
}

type UserOther struct {
	Age     int    `json:"age" validate:"required,min=0,max=100" example:"25"` // Required, user's age
	Country string `json:"country" validate:"required" example:"ru"`           // Required, user's country
}

type UserRegisterResponse struct {
	Token string `json:"token"` // JWT token
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email,min=8,max=120" example:"example@gmail.com"` // User's email, must be valid email address
	Password string `json:"password" validate:"required,password,min=8,max=60" example:"Password1234"` // User's password
}

type UserProfile struct {
	Email     string    `json:"email" validate:"required,email" example:"example@gmail.com"`                  // Required, email must be valid
	Name      string    `json:"name" validate:"required" example:"John"`                                      // Required, user's name
	Surname   string    `json:"surname" validate:"required" example:"Doe"`                                    // Required, user's surname
	AvatarURL string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg" validate:"url"` // User's avatar URL
	Other     UserOther `json:"other"`
}

type UserProfileUpdate struct {
	Name      *string `json:"name" example:"John" validate:"omitempty,omitnil,min=1,max=100"`                                          // Required, user's name
	Surname   *string `json:"surname" example:"Doe" validate:"omitempty,omitnil,min=1,max=120"`                                        // Required, user's surname
	AvatarURL *string `json:"avatar_url" example:"https://example.com/avatar.jpg" validate:"omitempty,omitnil,url,max=350"`            // User's avatar URL
	Password  *string `json:"password" validate:"omitempty,password" example:"Password1234" validate:"omitempty,omitnil,min=8,max=60"` // Required, password must meet certain requirements: must have upper case letters, lower case letters and digits
}

type Feed struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	SortBy string `query:"sort_by"`
	Active bool   `query:"active"`
}
