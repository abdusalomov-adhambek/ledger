package transfer

import "github.com/jackc/pgx/v5/pgxpool"

type TransferRepo struct {
	pxg *pgxpool.Pool
}

func NewTransferRepo(pxg *pgxpool.Pool) *TransferRepo {
	return &TransferRepo{pxg: pxg}
}
