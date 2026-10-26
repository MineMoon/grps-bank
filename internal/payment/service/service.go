package service

import (
	"context"
	"errors"
	"fmt"
	"grpc-bank/internal/payment/domain"
	accountpb "grpc-bank/pkg/api/account"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentRepository interface {
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error)
	CreatePending(ctx context.Context, tx *domain.Transaction) error
	UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus, errorMsg string) error
}

type PaymentService struct {
	repo PaymentRepository
	accountClient accountpb.AccountServiceClient
}

func NewPaymentService(repo PaymentRepository, accountClient accountpb.AccountServiceClient) *PaymentService {
	return &PaymentService{
		repo:          repo,
		accountClient: accountClient,
	}
}

func (s *PaymentService) ProcessPayment(
	ctx context.Context,
	idempotencyKey string,
	fromAccountID string,
	toAccountID string,
	amount decimal.Decimal,
) (*domain.Transaction, error) {
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotencyKey is required")
	}
	if fromAccountID == "" || toAccountID == "" {
		return nil, fmt.Errorf("sender and recipient accounts are required")
	}
	if fromAccountID == toAccountID {
		return nil, fmt.Errorf("sender account can't equal recipient account")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be strictly positive")
	}

	existingTx, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err == nil {
		switch existingTx.Status{
		case domain.StatusSuccess:
			return existingTx, nil
		case domain.StatusPending:
			return  nil, fmt.Errorf("a payment with this key is already being processed")
		case domain.StatusFailed:
			return nil, fmt.Errorf("the previous payment attempt was rejected: %w", err)
		}
	} else if !errors.Is(err, domain.ErrTransactionNotFound) {
		return nil, fmt.Errorf("idempotency check error: %w", err)
	}

	tx := &domain.Transaction{
		ID:             idempotencyKey,
		IdempotencyKey: idempotencyKey,
		FromAccountID:  fromAccountID,
		ToAccountID:    toAccountID,
		Amount:         amount,
		Status:         domain.StatusPending,
	}

	if err := s.repo.CreatePending(ctx, tx); err != nil {
		if errors.Is(err, domain.ErrDuplicateTransaction){
			return nil, fmt.Errorf("a request with this key has already been recorded")
		}
		return  nil, fmt. Errorf("failed to register transaction: %w", err)
	}

	transferReq := &accountpb.ExecuteTransferRequest{
		FromAccountId: fromAccountID,
		ToAccountId:   toAccountID,
		Amount:        amount.String(),
	}

	_, err = s.accountClient.ExecuteTransfer(ctx, transferReq)
	if err != nil {
		st, ok := status.FromError(err)
		var userErrMsg string

		if ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				userErrMsg = "Insufficient funds in the account"
			case codes.NotFound:
				userErrMsg = "One of the accounts was not found."
			default:
				userErrMsg = fmt.Sprintf("Banking core error: %s", st.Message())
			}
		} else {
			userErrMsg = fmt.Sprintf("Network communication failure: %v", err)
		}

		_ = s.repo.UpdateStatus(ctx, tx.ID, domain.StatusFailed, userErrMsg)
		tx.Status = domain.StatusFailed
		tx.ErrorMessage = userErrMsg

		return tx, errors.New(userErrMsg)
	}

	if err := s.repo.UpdateStatus(ctx, tx.ID, domain.StatusSuccess, ""); err != nil {
		return nil, fmt.Errorf("платеж проведен, но возникла ошибка обновления статуса: %w", err)	
	}

	tx.Status = domain.StatusSuccess
	return tx, nil
}

