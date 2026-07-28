package entry

import (
	"ledger/internal/domain/account"
	"ledger/internal/domain/transaction"
)

type Entry struct {
	id            string
	transactionId string
	accountId     string
	entryType     string
	amount        int64
	createdAt     string

	transaction *transaction.Transaction
	account     *account.Account
}

const (
	EntryTypeDebit  = "DEBIT"
	EntryTypeCredit = "CREDIT"
)

func NewEntry(id, transactionId, accountId, entryType string, amount int64) *Entry {
	return &Entry{
		id:            id,
		transactionId: transactionId,
		accountId:     accountId,
		entryType:     entryType,
		amount:        amount,
	}
}

func (e *Entry) ID() string {
	return e.id
}
func (e *Entry) TransactionID() string {
	return e.transactionId
}

func (e *Entry) AccountID() string {
	return e.accountId
}

func (e *Entry) EntryType() string {
	return e.entryType
}

func (e *Entry) Amount() int64 {
	return e.amount
}

func (e *Entry) CreatedAt() string {
	return e.createdAt
}

func (e *Entry) Transaction() *transaction.Transaction {
	return e.transaction
}

func (e *Entry) SetTransaction(transaction *transaction.Transaction) {
	e.transaction = transaction
}

func (e *Entry) Account() *account.Account {
	return e.account
}

func (e *Entry) SetAccount(account *account.Account) {
	e.account = account
}
