package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrDuplicateTransaction = errors.New("transaction with the same idempotency key already exists")
	ErrTransactionNotFound = errors.New("Transaction not found")
)

type TransactionStatus string 

const (
	StatusPending TransactionStatus = "PENDING"
	StatusSuccess TransactionStatus = "SUCCESS"
	StatusFailed TransactionStatus = "FAILED"
	
)

type Transaction struct {
	ID string
	IdempotencyKey string
	FromAccountID string
	ToAccountID string
	Amount decimal.Decimal
	Status TransactionStatus
	ErrorMessage string
	CreatedAt time.Time
}