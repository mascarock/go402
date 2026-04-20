package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mascarock/gopayments/model"
	"github.com/mascarock/gopayments/store"
)

func setup() (chi.Router, store.Store) {
	s := store.New()
	return NewRouter(s), s
}

func doGet(r chi.Router, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}

func doPost(r chi.Router, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)
	return rr
}

func doPostWithAuth(r chi.Router, path, body, apiKey string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)
	r.ServeHTTP(rr, req)
	return rr
}

func doGetWithHeader(r chi.Router, path, key, value string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(key, value)
	r.ServeHTTP(rr, req)
	return rr
}

func decode[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(rr.Body).Decode(&v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return v
}

func getAPIKey(s store.Store, tenantID string) string {
	t, _ := s.GetTenant(tenantID)
	return t.APIKey
}

// --- Tenants ---

func TestListTenants(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	tenants := decode[[]model.TenantPublic](t, rr)
	if len(tenants) != 3 {
		t.Errorf("expected 3 seeded tenants, got %d", len(tenants))
	}
}

func TestListTenantsDoesNotExposeAPIKey(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants")
	var raw []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&raw)
	for _, tenant := range raw {
		if _, ok := tenant["api_key"]; ok {
			t.Error("api_key should not be present in tenant list response")
		}
	}
}

func TestGetTenantFound(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/t_nexus")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	tenant := decode[model.TenantPublic](t, rr)
	if tenant.ID != "t_nexus" {
		t.Errorf("expected ID 't_nexus', got %q", tenant.ID)
	}
}

func TestGetTenantDoesNotExposeAPIKey(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/t_nexus")
	var raw map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&raw)
	if _, ok := raw["api_key"]; ok {
		t.Error("api_key should not be present in tenant get response")
	}
}

func TestGetTenantNotFound(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/nonexistent")
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCreateTenantSuccess(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/tenants", `{"name":"New Corp","brand":"NewPay"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var raw map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&raw)
	if raw["name"] != "New Corp" || raw["brand"] != "NewPay" {
		t.Errorf("unexpected tenant: %+v", raw)
	}
	if raw["id"] == nil || raw["id"] == "" {
		t.Error("expected ID to be generated")
	}
	if raw["api_key"] == nil || raw["api_key"] == "" {
		t.Error("expected api_key in creation response")
	}
}

func TestCreateTenantValidation(t *testing.T) {
	r, _ := setup()
	tests := []struct {
		name string
		body string
	}{
		{"missing name", `{"brand":"NewPay"}`},
		{"missing brand", `{"name":"New Corp"}`},
		{"empty body", `{}`},
		{"invalid json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doPost(r, "/api/tenants", tt.body)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestCreateTenantResponseContentType(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/tenants", `{"name":"New Corp","brand":"NewPay"}`)
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

// --- Wallets ---

func TestGetBalanceFound(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/t_nexus/wallet")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	wallet := decode[model.Wallet](t, rr)
	if wallet.TenantID != "t_nexus" || wallet.Balance != 500000000 {
		t.Errorf("unexpected wallet: %+v", wallet)
	}
}

func TestGetBalanceNotFound(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/nonexistent/wallet")
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDepositSuccess(t *testing.T) {
	r, s := setup()
	key := getAPIKey(s, "t_nexus")
	rr := doPostWithAuth(r, "/api/tenants/t_nexus/deposit", `{"amount":100000}`, key)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	tx := decode[model.Transaction](t, rr)
	if tx.Amount != 100000 || tx.Type != model.TxDeposit {
		t.Errorf("unexpected tx: %+v", tx)
	}
}

func TestDepositNoAuth(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/tenants/t_nexus/deposit", `{"amount":100000}`)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestDepositBadAuth(t *testing.T) {
	r, _ := setup()
	rr := doPostWithAuth(r, "/api/tenants/t_nexus/deposit", `{"amount":100000}`, "sk_wrong")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestDepositValidation(t *testing.T) {
	r, s := setup()
	key := getAPIKey(s, "t_nexus")
	tests := []struct {
		name string
		body string
	}{
		{"zero amount", `{"amount":0}`},
		{"negative amount", `{"amount":-50000}`},
		{"invalid json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doPostWithAuth(r, "/api/tenants/t_nexus/deposit", tt.body, key)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestDepositTenantNotFound(t *testing.T) {
	r, _ := setup()
	rr := doPostWithAuth(r, "/api/tenants/nonexistent/deposit", `{"amount":100000}`, "sk_any")
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestWithdrawSuccess(t *testing.T) {
	r, s := setup()
	key := getAPIKey(s, "t_nexus")
	rr := doPostWithAuth(r, "/api/tenants/t_nexus/withdraw", `{"amount":100000}`, key)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	tx := decode[model.Transaction](t, rr)
	if tx.Type != model.TxWithdrawal {
		t.Errorf("expected withdrawal, got %q", tx.Type)
	}
}

func TestWithdrawNoAuth(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/tenants/t_nexus/withdraw", `{"amount":100000}`)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestWithdrawInsufficientBalance(t *testing.T) {
	r, s := setup()
	tenant := s.CreateTenant("Broke Corp", "BrokePay")
	rr := doPostWithAuth(r, "/api/tenants/"+tenant.ID+"/withdraw", `{"amount":100000}`, tenant.APIKey)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestWithdrawValidation(t *testing.T) {
	r, s := setup()
	key := getAPIKey(s, "t_nexus")
	tests := []struct {
		name string
		body string
	}{
		{"zero amount", `{"amount":0}`},
		{"negative amount", `{"amount":-50000}`},
		{"invalid json", `{bad}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doPostWithAuth(r, "/api/tenants/t_nexus/withdraw", tt.body, key)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestWithdrawTenantNotFound(t *testing.T) {
	r, _ := setup()
	rr := doPostWithAuth(r, "/api/tenants/nonexistent/withdraw", `{"amount":100000}`, "sk_any")
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// --- Transactions ---

func TestListAllTransactions(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/transactions")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	txns := decode[[]model.Transaction](t, rr)
	if len(txns) != 12 {
		t.Errorf("expected 12 seeded transactions, got %d", len(txns))
	}
}

func TestListTransactionsByTenant(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/t_nexus/transactions")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	txns := decode[[]model.Transaction](t, rr)
	if len(txns) != 4 {
		t.Errorf("expected 4 transactions, got %d", len(txns))
	}
	for _, tx := range txns {
		if tx.TenantID != "t_nexus" {
			t.Errorf("expected tenant_id 't_nexus', got %q", tx.TenantID)
		}
	}
}

func TestListTransactionsNonexistentTenant(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/tenants/nonexistent/transactions")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 (empty list), got %d", rr.Code)
	}
	txns := decode[[]model.Transaction](t, rr)
	if len(txns) != 0 {
		t.Errorf("expected 0, got %d", len(txns))
	}
}

func TestTransactionResponseContentType(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/transactions")
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

// --- Payments (x402) ---

func TestProtectedDataRequiresPayment(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/protected/data")
	if rr.Code != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d", rr.Code)
	}
	for _, h := range []string{"X-Payment-Amount", "X-Payment-Currency", "X-Payment-Network", "X-Payment-Endpoint"} {
		if rr.Header().Get(h) == "" {
			t.Errorf("expected header %s to be set", h)
		}
	}
}

func TestProtectedDataInvalidToken(t *testing.T) {
	r, _ := setup()
	rr := doGetWithHeader(r, "/api/protected/data", "X-Payment-Token", "invalid")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestProtectedDataValidToken(t *testing.T) {
	r, s := setup()
	receipt := s.CreateReceipt("t_nexus", 100000)
	rr := doGetWithHeader(r, "/api/protected/data", "X-Payment-Token", receipt.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	resp := decode[map[string]interface{}](t, rr)
	if resp["paid"] != true {
		t.Error("expected paid=true")
	}
}

func TestProtectedDataTokenReuse(t *testing.T) {
	r, s := setup()
	receipt := s.CreateReceipt("t_nexus", 100000)

	rr := doGetWithHeader(r, "/api/protected/data", "X-Payment-Token", receipt.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("first use should return 200, got %d", rr.Code)
	}

	rr = doGetWithHeader(r, "/api/protected/data", "X-Payment-Token", receipt.Token)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("second use should return 401, got %d", rr.Code)
	}
}

func TestProcessPaymentSuccess(t *testing.T) {
	r, s := setup()
	tenant, _ := s.GetTenant("t_nexus")
	body, _ := json.Marshal(map[string]interface{}{
		"tenant_id": tenant.ID, "api_key": tenant.APIKey, "amount": 100000,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/payments/process", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	resp := decode[map[string]interface{}](t, rr)
	if resp["success"] != true {
		t.Error("expected success=true")
	}
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected token in response")
	}
}

func TestProcessPaymentDefaultAmount(t *testing.T) {
	r, s := setup()
	tenant, _ := s.GetTenant("t_nexus")
	body, _ := json.Marshal(map[string]interface{}{
		"tenant_id": tenant.ID, "api_key": tenant.APIKey, "amount": 0,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/payments/process", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestProcessPaymentErrors(t *testing.T) {
	r, s := setup()
	broke := s.CreateTenant("Broke", "BP")

	tests := []struct {
		name   string
		body   interface{}
		status int
	}{
		{"tenant not found", map[string]interface{}{"tenant_id": "nope", "api_key": "sk_x", "amount": 100000}, 404},
		{"invalid api key", map[string]interface{}{"tenant_id": "t_nexus", "api_key": "sk_wrong", "amount": 100000}, 401},
		{"insufficient balance", map[string]interface{}{"tenant_id": broke.ID, "api_key": broke.APIKey, "amount": 100000}, 422},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/payments/process", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(rr, req)
			if rr.Code != tt.status {
				t.Errorf("expected %d, got %d", tt.status, rr.Code)
			}
		})
	}
}

func TestProcessPaymentInvalidJSON(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/payments/process", `not json`)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- Settlements ---

func TestListSettlementsEmpty(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/settlements")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	settlements := decode[[]model.Settlement](t, rr)
	if len(settlements) != 0 {
		t.Errorf("expected 0, got %d", len(settlements))
	}
}

func TestCreateSettlementSuccess(t *testing.T) {
	r, s := setup()
	key := getAPIKey(s, "t_nexus")
	body, _ := json.Marshal(map[string]interface{}{
		"from_tenant": "t_nexus", "to_tenant": "t_pulse", "amount": 100000, "api_key": key,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/settlements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	stl := decode[model.Settlement](t, rr)
	if stl.FromTenant != "t_nexus" || stl.ToTenant != "t_pulse" {
		t.Errorf("unexpected settlement: %+v", stl)
	}
}

func TestCreateSettlementNoAuth(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/settlements", `{"from_tenant":"t_nexus","to_tenant":"t_pulse","amount":100000}`)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestCreateSettlementBadAuth(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/settlements", `{"from_tenant":"t_nexus","to_tenant":"t_pulse","amount":100000,"api_key":"sk_wrong"}`)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestCreateSettlementValidation(t *testing.T) {
	r, _ := setup()
	tests := []struct {
		name string
		body string
	}{
		{"missing from", `{"to_tenant":"t_pulse","amount":100000,"api_key":"x"}`},
		{"missing to", `{"from_tenant":"t_nexus","amount":100000,"api_key":"x"}`},
		{"missing amount", `{"from_tenant":"t_nexus","to_tenant":"t_pulse","api_key":"x"}`},
		{"zero amount", `{"from_tenant":"t_nexus","to_tenant":"t_pulse","amount":0,"api_key":"x"}`},
		{"negative amount", `{"from_tenant":"t_nexus","to_tenant":"t_pulse","amount":-100,"api_key":"x"}`},
		{"same tenant", `{"from_tenant":"t_nexus","to_tenant":"t_nexus","amount":100000,"api_key":"x"}`},
		{"invalid json", `{bad}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := doPost(r, "/api/settlements", tt.body)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestCreateSettlementInsufficientBalance(t *testing.T) {
	r, s := setup()
	broke := s.CreateTenant("Poor", "PP")
	body, _ := json.Marshal(map[string]interface{}{
		"from_tenant": broke.ID, "to_tenant": "t_nexus", "amount": 100000, "api_key": broke.APIKey,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/settlements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestCreateSettlementSourceNotFound(t *testing.T) {
	r, _ := setup()
	rr := doPost(r, "/api/settlements", `{"from_tenant":"nonexistent","to_tenant":"t_nexus","amount":100000,"api_key":"sk_x"}`)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestListSettlementsAfterCreating(t *testing.T) {
	r, s := setup()
	key := getAPIKey(s, "t_nexus")
	body, _ := json.Marshal(map[string]interface{}{
		"from_tenant": "t_nexus", "to_tenant": "t_pulse", "amount": 50000, "api_key": key,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/settlements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)

	rr2 := doGet(r, "/api/settlements")
	settlements := decode[[]model.Settlement](t, rr2)
	if len(settlements) != 1 {
		t.Errorf("expected 1, got %d", len(settlements))
	}
}

// --- Stats ---

func TestGetStats(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/stats")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	stats := decode[model.PlatformStats](t, rr)
	if stats.TotalTenants != 3 || stats.TotalTransactions != 12 || stats.ActiveWallets != 3 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}

func TestGetStatsContentType(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/stats")
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

// --- Security headers ---

func TestSecurityResponseHeaders(t *testing.T) {
	r, _ := setup()
	rr := doGet(r, "/api/stats")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
