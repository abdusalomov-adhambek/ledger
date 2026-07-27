package transfer

import "errors"

var (
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrSameAccount      = errors.New("same account")
	ErrCurrencyMismatch = errors.New("currency mismatch")
)
