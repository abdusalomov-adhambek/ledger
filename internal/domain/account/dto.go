package account

import (
	"time"
)

const (
	AccountStatusOpen    string = "OPEN"
	AccountStatusClosed  string = "CLOSED"
	AccountStatusBlocked string = "BLOCKED"
)

const (
	CurrencyUSD string = "USD"
	CurrencyEUR string = "EUR"
	CurrencyUZS string = "UZS"
)

type Account struct {
	id        string
	ownerId   string
	currency  string
	balance   int64
	status    string
	createdAt time.Time
	updatedAt time.Time
}

func NewAccount(
	id string,
	ownerId string,
	currency string,
	balance int64,
	status string,
) *Account {
	return &Account{
		id:       id,
		ownerId:  ownerId,
		currency: currency,
		balance:  balance,
		status:   status,
	}
}

func (a *Account) ID() string {
	return a.id
}

func (a *Account) SetID(id string) {
	a.id = id
}

func (a *Account) OwnerID() string {
	return a.ownerId
}

func (a *Account) Currency() string {
	return a.currency
}

func (a *Account) Balance() int64 {
	return a.balance
}

func (a *Account) Status() string {
	return a.status
}

func (a *Account) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Account) SetCreatedAt(createdAt time.Time) {
	a.createdAt = createdAt
}

func (a *Account) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *Account) SetUpdatedAt(updatedAt time.Time) {
	a.updatedAt = updatedAt
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if a.status != AccountStatusOpen {
		return ErrInvalidStatus
	}

	if a.balance < amount {
		return ErrInsufficientBalance
	}

	a.balance -= amount

	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if a.status != AccountStatusOpen {
		return ErrInvalidStatus
	}

	a.balance += amount

	return nil
}
