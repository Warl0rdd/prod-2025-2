package dto

type HTTPResponse struct {
	Status  string `json:"status"`                                     // HTTP error code
	Message string `json:"message,omitempty" example:"you are retard"` // Error message
}
