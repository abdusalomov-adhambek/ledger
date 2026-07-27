package account

import "errors"

var (
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInsufficientBalance = errors.New("insufficient balance")
)
