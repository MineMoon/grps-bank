package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound     = errors.New("account Not Found")
	ErrInsufficientFunds   = errors.New("insufficient Funds")
	ErrInvalidAmount       = errors.New("the transfer amount must be strictly positive")
	ErrSameAccountTransfer = errors.New("you cannot transfer funds to the same account")
	ErrAccountAlreadyExists = errors.New("the user already has an account in this currency")
)

type Account struct {
	ID            string
	UserID        string
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
