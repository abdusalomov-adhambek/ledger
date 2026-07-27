package account

type CreateAccountRequest struct {
	OwnerID  string `json:"owner_id" validate:"required"`
	Currency string `json:"currency" validate:"required"`
	Balance  int64  `json:"balance" validate:"required"`
}

type CreateAccountResponse struct {
	ID string `json:"id"`
}

type GetByIDRequest struct {
	ID string `json:"id"`
}

type GetByIDResponse struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	Balance   int64  `json:"balance"`
	CreatedAt string `json:"created_at"`
}
