package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mascarock/gopayments/store"
)

func NewRouter(s store.Store, middlewares ...func(http.Handler) http.Handler) chi.Router {
	h := New(s)
	r := chi.NewRouter()
	for _, mw := range middlewares {
		r.Use(mw)
	}

	r.Route("/api", func(r chi.Router) {
		r.Get("/stats", h.GetStats)
		r.Get("/transactions", h.ListAllTransactions)

		r.Route("/tenants", func(r chi.Router) {
			r.Get("/", h.ListTenants)
			r.Post("/", h.CreateTenant)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetTenant)
				r.Get("/credentials", h.GetCredentials)
				r.Get("/wallet", h.GetBalance)
				r.Post("/deposit", h.Deposit)
				r.Post("/withdraw", h.Withdraw)
				r.Get("/transactions", h.ListTransactionsByTenant)
			})
		})

		r.Route("/settlements", func(r chi.Router) {
			r.Get("/", h.ListSettlements)
			r.Post("/", h.CreateSettlement)
		})

		r.Route("/payments", func(r chi.Router) {
			r.Post("/process", h.ProcessPayment)
		})

		r.Get("/protected/data", h.ProtectedData)
	})

	return r
}
