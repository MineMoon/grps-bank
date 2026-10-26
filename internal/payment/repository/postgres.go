package repository

import (
	"context"
	"errors"
	"fmt"
	"grpc-bank/internal/payment/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount, status, error_message, created_at 
		FROM transactions
		WHERE idempotency_key = $1;
	`
	var tx domain.Transaction
	var errMsg *string

	err := r.pool.QueryRow(ctx, query, key).Scan(
		&tx.ID,
		&tx.IdempotencyKey,
		&tx.FromAccountID,
		&tx.ToAccountID,
		&tx.Amount,
		&tx.Status,
		&errMsg,
		&tx.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("transaction lookup failed: %w", err)
	}

	if errMsg != nil {
		tx.ErrorMessage = *errMsg
	}

	return &tx, nil
}

func (r *PaymentRepository) CreatePending(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(ctx, query,
		tx.IdempotencyKey,
		tx.FromAccountID, 
		tx.ToAccountID,
		tx.Amount,
		domain.StatusPending,
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrDuplicateTransaction
		}
		return fmt.Errorf("error creating payment record: %w", err)
	}

	tx.Status = domain.StatusPending
	return nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus, errorMsg string) error {
	query := `
		UPDATE transactions 
		SET status = $1, error_message = $2
		WHERE id = $3;
	`
	var errStr *string
	if errorMsg != "" {
		errStr = &errorMsg
	}

	_, err := r.pool.Exec(ctx, query, status, errStr, id)
	if err != nil {
		return fmt.Errorf("update status error: %w", err)
	}
	return nil
}

