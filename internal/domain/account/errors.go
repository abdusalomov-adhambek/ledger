package account

import "errors"

var (
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidStatus       = errors.New("invalid account status")
	ErrInsufficientBalance = errors.New("insufficient balance")
)
