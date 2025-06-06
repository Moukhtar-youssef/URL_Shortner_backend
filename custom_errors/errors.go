package customerrors

import "errors"

var (
	ErrAlreadyShortened = errors.New("this URL is already shortened")
	ErrEmptyLongURL     = errors.New("Long urls can't be empty")
	ErrInvalidLongURL   = errors.New("Invalid long url")
)
