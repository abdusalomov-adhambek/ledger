package grpc

import (
	"context"
	ledgerv1 "ledger/api/proto"
	"ledger/internal/application/account"
	"ledger/internal/application/entry"
	"ledger/internal/application/transaction"
	"ledger/internal/application/transfer"
	"log"

	entryDomain "ledger/internal/domain/entry"

	"google.golang.org/grpc/metadata"
)

type LedgerService struct {
	ledgerv1.UnimplementedLedgerServiceServer
	accountApp     *account.AccountApplication
	transferApp    *transfer.TransferApp
	transactionApp *transaction.TransactionApplication
	entryApp       *entry.EntryApplication
}

func NewLedgerGRPCService(
	accountApp *account.AccountApplication,
	transferApp *transfer.TransferApp,
	transactionApp *transaction.TransactionApplication,
	entryApp *entry.EntryApplication) *LedgerService {
	return &LedgerService{
		accountApp:     accountApp,
		transferApp:    transferApp,
		transactionApp: transactionApp,
		entryApp:       entryApp,
	}
}

func (s *LedgerService) CreateAccount(ctx context.Context, req *ledgerv1.CreateAccountRequest) (*ledgerv1.CreateAccountResponse, error) {
	currency := req.Currency.String()

	res, err := s.accountApp.CreateAccount(ctx, account.CreateAccountRequest{
		OwnerID:  req.OwnerId,
		Currency: currency,
		Balance:  req.Balance,
	})
	if err != nil {
		return nil, err
	}

	return &ledgerv1.CreateAccountResponse{
		Id: res.ID,
	}, nil
}

func (s *LedgerService) GetAccount(ctx context.Context, req *ledgerv1.GetAccountRequest) (*ledgerv1.GetAccountResponse, error) {
	res, err := s.accountApp.GetByID(ctx, account.GetByIDRequest{
		ID: req.Id,
	})
	if err != nil {
		return nil, err
	}

	return &ledgerv1.GetAccountResponse{
		Id:        res.ID,
		OwnerId:   res.OwnerID,
		Currency:  res.Currency,
		Balance:   res.Balance,
		Status:    res.Status,
		CreatedAt: res.CreatedAt,
	}, nil
}

func (s *LedgerService) Transfer(ctx context.Context, req *ledgerv1.TransferRequest) (*ledgerv1.TransferResponse, error) {

	var idempotencyKey string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("idempotency-key")
		if len(values) > 0 {
			idempotencyKey = values[0]
		}
	}
	log.Println("idempotencyKey", idempotencyKey)
	res, err := s.transferApp.Transfer(ctx, &transfer.TransferRequest{
		FromAccountID:  req.FromAccountId,
		ToAccountID:    req.ToAccountId,
		Amount:         req.Amount,
		Description:    req.Description,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, err
	}

	return &ledgerv1.TransferResponse{
		Id: res,
	}, nil
}

func (s *LedgerService) GetEntryHistory(ctx context.Context, req *ledgerv1.GetEntryHistoryRequest) (*ledgerv1.GetEntryHistoryResponse, error) {
	res, count, err := s.entryApp.GetHistoryList(ctx, &entryDomain.Filter{
		Limit:  &req.Limit,
		Offset: &req.Offset,
		Page:   &req.Page,
	}, &entry.GetHistoryListRequest{
		AccountID: req.AccountId,
	})
	if err != nil {
		return nil, err
	}

	var data []*ledgerv1.GentryHistory
	for _, entry := range res {
		log.Println("entry", entry.ID)
		data = append(data, &ledgerv1.GentryHistory{
			Id:            entry.ID,
			TransactionId: entry.TransactionID,
			Amount:        entry.Amount,
			Description:   entry.Description,
			Status:        entry.Status,
			EntryType:     entry.EntryType,
			CreatedAt:     entry.CreatedAt,
		})
	}
	return &ledgerv1.GetEntryHistoryResponse{
		Data:  data,
		Count: count,
	}, nil
}

func (s *LedgerService) ReverseTransfer(ctx context.Context, req *ledgerv1.ReverseTransferRequest) (*ledgerv1.ReverseTransferResponse, error) {
	if err := s.transactionApp.ReverseTransaction(ctx, &transaction.ReverseTransactionRequest{
		TransactionID: req.TransactionId,
	}); err != nil {
		return nil, err
	}

	return &ledgerv1.ReverseTransferResponse{
		Message: "Transaction reversed successfully",
	}, nil
}
