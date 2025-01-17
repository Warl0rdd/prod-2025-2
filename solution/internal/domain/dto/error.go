package dto

type HTTPError struct {
	Status  string `json:"status"`                           // HTTP error code
	Message string `json:"message" example:"you are retard"` // Error message
}
