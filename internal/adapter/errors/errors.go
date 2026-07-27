package errors

import "errors"

var (
	ErrInvalidCurrency = errors.New("invalid currency")
	ErrInvalidAmount   = errors.New("invalid amount")
)
