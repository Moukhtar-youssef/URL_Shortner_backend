package customerrors

import "errors"

var (
	ErrAlreadyShortened = errors.New("this URL is already shortened")
	ErrInvalidLongURL   = errors.New("Invalid long url")
)
