package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mascarock/gopayments/store"
)

type Handler struct {
	store store.Store
}

func New(s store.Store) *Handler {
	return &Handler{store: s}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) tenantID(r *http.Request) string {
	return chi.URLParam(r, "id")
}

// authenticateTenant extracts API key from Authorization header, looks up the
// tenant by the {id} URL param, and verifies the key using constant-time
// comparison. Returns the tenant ID on success.
func (h *Handler) authenticateTenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := h.tenantID(r)
	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "missing Authorization header")
		return "", false
	}

	tenant, ok := h.store.GetTenant(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return "", false
	}

	if subtle.ConstantTimeCompare([]byte(tenant.APIKey), []byte(apiKey)) != 1 {
		writeError(w, http.StatusUnauthorized, "invalid API key")
		return "", false
	}
	return id, true
}

// --- Tenants ---

func (h *Handler) GetCredentials(w http.ResponseWriter, r *http.Request) {
	id := h.tenantID(r)
	tenant, ok := h.store.GetTenant(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id": tenant.ID,
		"api_key":   tenant.APIKey,
	})
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants := h.store.ListTenants()
	public := make([]interface{}, len(tenants))
	for i, t := range tenants {
		p := t.Public()
		public[i] = p
	}
	writeJSON(w, http.StatusOK, public)
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenant, ok := h.store.GetTenant(h.tenantID(r))
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	writeJSON(w, http.StatusOK, tenant.Public())
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Brand string `json:"brand"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Name == "" || req.Brand == "" {
		writeError(w, http.StatusBadRequest, "name and brand are required")
		return
	}
	tenant := h.store.CreateTenant(req.Name, req.Brand)
	// Return full tenant (with API key) only on creation so the caller can store it
	writeJSON(w, http.StatusCreated, struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Brand     string `json:"brand"`
		APIKey    string `json:"api_key"`
		CreatedAt string `json:"created_at"`
	}{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Brand:     tenant.Brand,
		APIKey:    tenant.APIKey,
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// --- Wallets ---

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	wallet, ok := h.store.GetWallet(h.tenantID(r))
	if !ok {
		writeError(w, http.StatusNotFound, "wallet not found")
		return
	}
	writeJSON(w, http.StatusOK, wallet)
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	id, ok := h.authenticateTenant(w, r)
	if !ok {
		return
	}
	amount, ok := h.parseAmount(w, r)
	if !ok {
		return
	}
	tx, err := h.store.Deposit(id, amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	id, ok := h.authenticateTenant(w, r)
	if !ok {
		return
	}
	amount, ok := h.parseAmount(w, r)
	if !ok {
		return
	}
	tx, err := h.store.Withdraw(id, amount)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) parseAmount(w http.ResponseWriter, r *http.Request) (int64, bool) {
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return 0, false
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return 0, false
	}
	return req.Amount, true
}

// --- Transactions ---

func (h *Handler) ListTransactionsByTenant(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetTransactions(h.tenantID(r)))
}

func (h *Handler) ListAllTransactions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetAllTransactions())
}

// --- Payments (x402) ---

const MicropaymentAmount int64 = 100000

func (h *Handler) ProtectedData(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Payment-Token")
	if token == "" {
		w.Header().Set("X-Payment-Amount", "100000")
		w.Header().Set("X-Payment-Currency", "USD")
		w.Header().Set("X-Payment-Network", "gopayments-sim")
		w.Header().Set("X-Payment-Endpoint", "/api/payments/process")
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":    "payment required",
			"amount":   MicropaymentAmount,
			"currency": "USD",
			"message":  "This endpoint requires a micropayment. Submit payment via /api/payments/process and retry with X-Payment-Token header.",
		})
		return
	}

	receipt, valid := h.store.ValidateReceipt(token)
	if !valid {
		writeError(w, http.StatusUnauthorized, "invalid or used payment token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    "Premium market data: BTC/USD $67,432.18 | ETH/USD $3,891.42 | SOL/USD $178.93",
		"paid":    true,
		"receipt": receipt.Token,
		"cost":    receipt.Amount,
	})
}

func (h *Handler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenant_id"`
		APIKey   string `json:"api_key"`
		Amount   int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	tenant, ok := h.store.GetTenant(req.TenantID)
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if subtle.ConstantTimeCompare([]byte(tenant.APIKey), []byte(req.APIKey)) != 1 {
		writeError(w, http.StatusUnauthorized, "invalid API key")
		return
	}

	amount := req.Amount
	if amount <= 0 {
		amount = MicropaymentAmount
	}

	_, err := h.store.ProcessPayment(req.TenantID, amount, map[string]string{
		"type":     "x402_micropayment",
		"endpoint": "/api/protected/data",
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	receipt := h.store.CreateReceipt(req.TenantID, amount)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   receipt.Token,
		"amount":  receipt.Amount,
		"message": "Payment processed. Use token in X-Payment-Token header to access protected resource.",
	})
}

// --- Settlements ---

func (h *Handler) CreateSettlement(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromTenant string `json:"from_tenant"`
		ToTenant   string `json:"to_tenant"`
		Amount     int64  `json:"amount"`
		APIKey     string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.FromTenant == "" || req.ToTenant == "" || req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "from_tenant, to_tenant, and positive amount required")
		return
	}
	if req.FromTenant == req.ToTenant {
		writeError(w, http.StatusBadRequest, "cannot settle to same tenant")
		return
	}

	fromTenant, ok := h.store.GetTenant(req.FromTenant)
	if !ok {
		writeError(w, http.StatusNotFound, "source tenant not found")
		return
	}
	if subtle.ConstantTimeCompare([]byte(fromTenant.APIKey), []byte(req.APIKey)) != 1 {
		writeError(w, http.StatusUnauthorized, "invalid API key for source tenant")
		return
	}

	settlement, err := h.store.CreateSettlement(req.FromTenant, req.ToTenant, req.Amount)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, settlement)
}

func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.ListSettlements())
}

// --- Stats ---

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetStats())
}
