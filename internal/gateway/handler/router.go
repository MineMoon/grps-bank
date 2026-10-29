package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *GatewayHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/accounts", h.CreateAccount)
		r.Get("/accounts/{id}", h.GetAccount)
		r.Post("/accounts/deposit", h.Deposit)
		r.Post("/transfers", h.CreateTransfer)
	})

	return r
}