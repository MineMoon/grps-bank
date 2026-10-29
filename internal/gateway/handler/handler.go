package handler

import (
	"encoding/json"
	accountpb "grpc-bank/pkg/api/account"
	paymentpb "grpc-bank/pkg/api/payment"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayHandler struct {
	accountClient accountpb.AccountServiceClient
	paymentClient paymentpb.PaymentServiceClient
}

func NewGatewayHandler(accountClient accountpb.AccountServiceClient, paymentClient paymentpb.PaymentServiceClient) *GatewayHandler {
	return &GatewayHandler{
		accountClient: accountClient,
		paymentClient: paymentClient,
	}
}

func (h *GatewayHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.accountClient.CreateAccount(r.Context(), &accountpb.CreateAccountRequest{
		UserId:   req.UserId,
		Currency: req.Currency,
	})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	resp := CreateAccountResponse{
		AccountId:     grpcResp.GetAccountId(),
		AccountNumber: grpcResp.GetAccountNumber(),
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		http.Error(w, "account id is requried", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.accountClient.GetAccount(r.Context(), &accountpb.GetAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	resp := AccountResponse{
		AccountId:     grpcResp.GetAccountId(),
		UserId:        grpcResp.GetUserId(),
		AccountNumber: grpcResp.GetAccountNumber(),
		Balance:       grpcResp.GetBalance(),
		Currency:      grpcResp.GetCurrency(),
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.accountClient.Deposit(r.Context(), &accountpb.DepositRequest{
		AccountId: req.AccountId,
		Amount:    req.Amount,
	})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	resp := DepositResponse{
		Status:       grpcResp.Status,
		BalanceAfter: grpcResp.BalanceAfter,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadGateway)
		return
	}

	grpcResp, err := h.paymentClient.CreatePayment(r.Context(), &paymentpb.CreatePaymentRequest{
		IdempotencyKey: req.IdempotencyKey,
		FromAccountId:  req.FromAccountId,
		ToAccountId:    req.ToAccountId,
		Amount:         req.Amount,
	})

	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(grpcResp)
}


func handleGRPCErr(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch st.Code() {
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.AlreadyExists:
		http.Error(w, st.Message(), http.StatusConflict)
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.FailedPrecondition:
		http.Error(w, st.Message(), http.StatusUnprocessableEntity)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

