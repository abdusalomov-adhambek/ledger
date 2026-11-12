package transaction

import "errors"

var (
	ErrNoEntriesFound  = errors.New("no entries found for transaction")
	ErrAlreadyReversed = errors.New("transaction already reversed or is itself a reversal")
)
