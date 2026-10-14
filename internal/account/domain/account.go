package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound     = errors.New("Account Not Found")
	ErrInsufficientFunds   = errors.New("Insufficient Funds")
	ErrInvalidAmount       = errors.New("The transfer amount must be strictly positive")
	ErrSameAccountTransfer = errors.New("You cannot transfer funds to the same account")
)

type Account struct {
	ID            string
	UserId        string
	AccountNumber string
	Balance       decimal.Decimal
	Currency      string
	CreatedAt     time.Time
}

type Card struct {
	ID         string
	AccountID  string
	CardNumber string
}
