package entry

type GetHistoryListRequest struct {
	AccountID string `json:"account_id"`
}

type GetListResponse struct {
	ID            string `json:"id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	EntryType     string `json:"entry_type"`
}
