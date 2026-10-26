package handler

import (
	"context"
	"time"

	"grpc-bank/internal/payment/service"
	paymentpb "grpc-bank/pkg/api/payment"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *paymentpb.CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	if req.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "idempoteny_key is requried")
	}

	if req.GetToAccountId() == "" || req.GetFromAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sender and recipient accounts are required")
	}

	amount, err := decimal.NewFromString(req.GetAmount())
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "incorrect transfer amount")
	}

	tx, err := h.svc.ProcessPayment(
		ctx,
		req.GetIdempotencyKey(),
		req.GetFromAccountId(),
		req.GetToAccountId(),
		amount,
	)
	if err != nil {
		return  nil, status.Errorf(codes.FailedPrecondition, "failed payment: %v", err)
	}

	return &paymentpb.CreatePaymentResponse{
		PaymentId: tx.ID,
		Status:    string(tx.Status),
		CreatedAt: tx.CreatedAt.Format(time.RFC3339),
	}, nil
}