package transfer

type Transfer struct {
	fromAccountId string
	toAccountId   string
	amount        int64
}

func (t *Transfer) NewTransfer(
	fromAccountId string,
	toAccountId string,
	amount int64,
) *Transfer {
	return &Transfer{
		fromAccountId: fromAccountId,
		toAccountId:   toAccountId,
		amount:        amount,
	}
}

func (t *Transfer) FromAccountId() string {
	return t.fromAccountId
}

func (t *Transfer) ToAccountId() string {
	return t.toAccountId
}

func (t *Transfer) Amount() int64 {
	return t.amount
}
