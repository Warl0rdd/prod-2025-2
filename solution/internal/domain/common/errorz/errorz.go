package errorz

import "errors"

var (
	AuthHeaderIsEmpty = errors.New("auth header is empty")
	Forbidden         = errors.New("forbidden")
)
