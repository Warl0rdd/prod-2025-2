package errorz

import "errors"

var (
	AuthHeaderIsEmpty = errors.New("auth header is empty")
	Forbidden         = errors.New("forbidden")
	NotFound          = errors.New("not found")
	EmailTaken        = errors.New("email already taken")
)
