package handler

import (
	"context"
	"errors"

	"grpc-bank/internal/account/domain"
	"grpc-bank/internal/account/service"
	accountpb "grpc-bank/pkg/api/account"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountHandler struct {
	accountpb.UnimplementedAccountServiceServer
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) GetAccount(ctx context.Context, req *accountpb.GetAccountRequest) (*accountpb.GetAccountResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	acc, err := h.svc.GetAccount(ctx, req.AccountId)

	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "Internal Server Errror: %v", err)
	}

	return &accountpb.GetAccountResponse{
		AccountId:     acc.ID,
		UserId:        acc.UserId,
		AccountNumber: acc.AccountNumber,
		Balance:       acc.Balance.String(),
		Currency:      acc.Currency,
	}, nil

}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *accountpb.CreateAccountRequest) (*accountpb.CreateAccountResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	acc, err := h.svc.CreateAccount(ctx, req.GetUserId(), req.GetCurrency())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create an account: %v", err)
	}

	return &accountpb.CreateAccountResponse{
		AccountId:     acc.ID,
		AccountNumber: acc.AccountNumber,
	}, nil
}

func (h *AccountHandler) ExecuteTransfer(ctx context.Context, req *accountpb.ExecuteTransferRequest) (*accountpb.ExecuteTransferResponse, error) {
	if req.GetFromAccountId() == "" || req.GetToAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	amount, err := decimal.NewFromString(req.GetAmount())
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "incorrect amount")
	}

	newBalance, err := h.svc.Transfer(ctx, req.GetFromAccountId(), req.GetToAccountId(), amount)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountNotFound):
			return nil, status.Error(codes.NotFound, "one of the accounts was not found")
		case errors.Is(err, domain.ErrInsufficientFunds):
			return nil, status.Error(codes.FailedPrecondition, "insufficient funds in the account")
		case errors.Is(err, domain.ErrSameAccountTransfer):
			return nil, status.Error(codes.InvalidArgument, "cannot be transferred to the same account")
		default:
			return nil, status.Errorf(codes.Internal, "transaction processing error: %v", err)
		}
	}

	return &accountpb.ExecuteTransferResponse{
		Status:           "SUCCESS",
		FromBalanceAfter: newBalance.String(),
	}, nil
}

func (h *AccountHandler) Deposit(ctx context.Context, req *accountpb.DepositRequest) (*accountpb.DepositResponse, error) {

	amount, err := decimal.NewFromString(req.GetAmount())
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "Incorrect top-up amount")
	}

	newBalance, err := h.svc.Deposit(ctx, req.GetAccountId(), amount)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "deposit error: %v", err)
	}

	return &accountpb.DepositResponse{
		Status:       "SUCCESS",
		BalanceAfter: newBalance.String(),
	}, nil
}
