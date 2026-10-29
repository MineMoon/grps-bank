package repository

import (
	"context"
	"errors"
	"fmt"
	"grpc-bank/internal/account/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type AccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountReposiroty(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

func (r *AccountRepository) GetbyID(ctx context.Context, id string) (*domain.Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, created_at
			FROM accounts
			WHERE id = $1;
	`

	var acc domain.Account
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.AccountNumber,
		&acc.Balance,
		&acc.Currency,
		&acc.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("request execution error: %w", err)
	}

	return &acc, nil
}

func (r *AccountRepository) Create(ctx context.Context, acc *domain.Account) error {
	query := `
		INSERT INTO accounts (user_id, account_number, balance, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`
	err := r.pool.QueryRow(ctx, query, acc.UserID, acc.AccountNumber, acc.Balance, acc.Currency).
		Scan(&acc.ID, &acc.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrAccountAlreadyExists
		}
		return fmt.Errorf("account Insert Error: %w", err)
	}
	return nil
}


func (r *AccountRepository) Transfer(ctx context.Context, fromID, toID string, amount decimal.Decimal) (decimal.Decimal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	firstID, secondID := fromID, toID
	if fromID > toID {
		firstID, secondID = toID, fromID
	}

	lockQuery := `
		SELECT id, balance FROM accounts WHERE id IN ($1, $2) ORDER BY id FOR UPDATE;
	`
	rows, err := tx.Query(ctx, lockQuery, firstID, secondID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("account blocking error (FOR UPDATE): %w", err)
	}
	defer rows.Close()

	balances := make(map[string]decimal.Decimal)
	for rows.Next() {
		var id string
		var bal decimal.Decimal
		if err := rows.Scan(&id, &bal); err != nil {
			return decimal.Zero, err
		}
		balances[id] = bal
	}

	fromBalance, existsFrom := balances[fromID]
	_, existsTo := balances[toID]
	if !existsFrom || !existsTo {
		return decimal.Zero, domain.ErrAccountNotFound
	}

	if fromBalance.LessThan(amount) {
		return decimal.Zero, domain.ErrInsufficientFunds
	}

	newFromBalance := fromBalance.Sub(amount)
	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2;`, amount, fromID)

	if err != nil {
		return decimal.Zero, fmt.Errorf("write-off error: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2;`, amount, toID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("crediting error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return decimal.Zero, fmt.Errorf("transaction Commit Error: %w", err)
	}

	return newFromBalance, nil
}

func (r *AccountRepository) Deposit(ctx context.Context, accountID string, amount decimal.Decimal) (decimal.Decimal, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, domain.ErrInvalidAmount
	}

	query := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		RETURNING 	balance;
	`

	var newBalance decimal.Decimal
	err := r.pool.QueryRow(ctx, query, amount, accountID).Scan(&newBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, domain.ErrAccountNotFound
		}
		return decimal.Zero, fmt.Errorf("Deposit Account Error: %w", err)
	}

	return newBalance, nil
}
