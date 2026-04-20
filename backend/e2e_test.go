package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mascarock/gopayments/handler"
	"github.com/mascarock/gopayments/model"
	"github.com/mascarock/gopayments/store"
)

func setupE2EServer() (*httptest.Server, store.Store) {
	s := store.New()
	return httptest.NewServer(handler.NewRouter(s)), s
}

func postWithAuth(url, body, apiKey string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", apiKey)
	}
	return http.DefaultClient.Do(req)
}

func TestE2E_FullPaymentFlow(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/tenants", "application/json",
		bytes.NewBufferString(`{"name":"E2E Corp","brand":"E2EPay"}`))
	if err != nil {
		t.Fatalf("create tenant request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 creating tenant, got %d", resp.StatusCode)
	}

	var createResp struct {
		ID     string `json:"id"`
		APIKey string `json:"api_key"`
	}
	json.NewDecoder(resp.Body).Decode(&createResp)
	if createResp.ID == "" || createResp.APIKey == "" {
		t.Fatal("expected tenant ID and API key in creation response")
	}
	tenantID := createResp.ID
	apiKey := createResp.APIKey

	resp, err = http.Get(srv.URL + "/api/tenants/" + tenantID)
	if err != nil {
		t.Fatalf("get tenant request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/api/tenants/" + tenantID + "/wallet")
	if err != nil {
		t.Fatalf("get wallet request failed: %v", err)
	}
	defer resp.Body.Close()
	var wallet model.Wallet
	json.NewDecoder(resp.Body).Decode(&wallet)
	if wallet.Balance != 0 {
		t.Fatalf("expected initial balance 0, got %d", wallet.Balance)
	}

	resp, err = postWithAuth(srv.URL+"/api/tenants/"+tenantID+"/deposit", `{"amount":1000000}`, apiKey)
	if err != nil {
		t.Fatalf("deposit request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for deposit, got %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/api/tenants/" + tenantID + "/wallet")
	if err != nil {
		t.Fatalf("get wallet failed: %v", err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&wallet)
	if wallet.Balance != 1000000 {
		t.Fatalf("expected balance 1000000, got %d", wallet.Balance)
	}

	resp, err = postWithAuth(srv.URL+"/api/tenants/"+tenantID+"/withdraw", `{"amount":300000}`, apiKey)
	if err != nil {
		t.Fatalf("withdraw request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for withdraw, got %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/api/tenants/" + tenantID + "/wallet")
	if err != nil {
		t.Fatalf("get wallet failed: %v", err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&wallet)
	if wallet.Balance != 700000 {
		t.Fatalf("expected balance 700000, got %d", wallet.Balance)
	}

	resp, err = http.Get(srv.URL + "/api/tenants/" + tenantID + "/transactions")
	if err != nil {
		t.Fatalf("get transactions failed: %v", err)
	}
	defer resp.Body.Close()
	var txns []*model.Transaction
	json.NewDecoder(resp.Body).Decode(&txns)
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}
}

func TestE2E_X402MicropaymentFlow(t *testing.T) {
	srv, s := setupE2EServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/protected/data")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", resp.StatusCode)
	}

	tenant, _ := s.GetTenant("t_nexus")

	paymentBody, _ := json.Marshal(map[string]interface{}{
		"tenant_id": tenant.ID,
		"api_key":   tenant.APIKey,
		"amount":    100000,
	})
	resp, err = http.Post(srv.URL+"/api/payments/process", "application/json", bytes.NewReader(paymentBody))
	if err != nil {
		t.Fatalf("payment request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for payment, got %d", resp.StatusCode)
	}

	var paymentResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&paymentResp)
	token, ok := paymentResp["token"].(string)
	if !ok || token == "" {
		t.Fatal("expected token in payment response")
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/protected/data", nil)
	req.Header.Set("X-Payment-Token", token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with valid token, got %d", resp.StatusCode)
	}

	var dataResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&dataResp)
	if dataResp["paid"] != true {
		t.Error("expected paid=true")
	}

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/api/protected/data", nil)
	req.Header.Set("X-Payment-Token", token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for reused token, got %d", resp.StatusCode)
	}
}

func TestE2E_SettlementFlow(t *testing.T) {
	srv, s := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/tenants/t_nexus/wallet")
	var nexusWallet model.Wallet
	json.NewDecoder(resp.Body).Decode(&nexusWallet)
	resp.Body.Close()
	initialNexus := nexusWallet.Balance

	resp, _ = http.Get(srv.URL + "/api/tenants/t_pulse/wallet")
	var pulseWallet model.Wallet
	json.NewDecoder(resp.Body).Decode(&pulseWallet)
	resp.Body.Close()
	initialPulse := pulseWallet.Balance

	nexusTenant, _ := s.GetTenant("t_nexus")
	body, _ := json.Marshal(map[string]interface{}{
		"from_tenant": "t_nexus",
		"to_tenant":   "t_pulse",
		"amount":      100000000,
		"api_key":     nexusTenant.APIKey,
	})
	resp, _ = http.Post(srv.URL+"/api/settlements", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(srv.URL + "/api/tenants/t_nexus/wallet")
	json.NewDecoder(resp.Body).Decode(&nexusWallet)
	resp.Body.Close()
	if nexusWallet.Balance != initialNexus-100000000 {
		t.Errorf("expected nexus balance %d, got %d", initialNexus-100000000, nexusWallet.Balance)
	}

	resp, _ = http.Get(srv.URL + "/api/tenants/t_pulse/wallet")
	json.NewDecoder(resp.Body).Decode(&pulseWallet)
	resp.Body.Close()
	if pulseWallet.Balance != initialPulse+100000000 {
		t.Errorf("expected pulse balance %d, got %d", initialPulse+100000000, pulseWallet.Balance)
	}

	resp, _ = http.Get(srv.URL + "/api/settlements")
	var settlements []*model.Settlement
	json.NewDecoder(resp.Body).Decode(&settlements)
	resp.Body.Close()
	if len(settlements) != 1 {
		t.Errorf("expected 1 settlement, got %d", len(settlements))
	}
}

func TestE2E_DepositWithoutAuth(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/tenants/t_nexus/deposit", "application/json",
		bytes.NewBufferString(`{"amount":100000}`))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_DepositWithWrongKey(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := postWithAuth(srv.URL+"/api/tenants/t_nexus/deposit", `{"amount":100000}`, "sk_wrong")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong key, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_OverdraftRejected(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/tenants", "application/json",
		bytes.NewBufferString(`{"name":"Broke Corp","brand":"BrokePay"}`))
	var createResp struct {
		ID     string `json:"id"`
		APIKey string `json:"api_key"`
	}
	json.NewDecoder(resp.Body).Decode(&createResp)
	resp.Body.Close()

	resp, _ = postWithAuth(srv.URL+"/api/tenants/"+createResp.ID+"/withdraw", `{"amount":100000}`, createResp.APIKey)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for overdraft, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(srv.URL + "/api/tenants/" + createResp.ID + "/wallet")
	var wallet model.Wallet
	json.NewDecoder(resp.Body).Decode(&wallet)
	resp.Body.Close()
	if wallet.Balance != 0 {
		t.Errorf("expected balance 0, got %d", wallet.Balance)
	}
}

func TestE2E_StatsReflectOperations(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/stats")
	var stats model.PlatformStats
	json.NewDecoder(resp.Body).Decode(&stats)
	resp.Body.Close()
	if stats.TotalTenants != 3 {
		t.Errorf("expected 3 tenants, got %d", stats.TotalTenants)
	}

	resp, _ = http.Post(srv.URL+"/api/tenants", "application/json",
		bytes.NewBufferString(`{"name":"Stats Corp","brand":"StatsPay"}`))
	resp.Body.Close()

	resp, _ = http.Get(srv.URL + "/api/stats")
	json.NewDecoder(resp.Body).Decode(&stats)
	resp.Body.Close()
	if stats.TotalTenants != 4 {
		t.Errorf("expected 4 tenants after creation, got %d", stats.TotalTenants)
	}
}

func TestE2E_TenantNotFound(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/tenants/nonexistent")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_InvalidCreateTenant(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"missing brand", `{"name":"Corp"}`},
		{"missing name", `{"brand":"Pay"}`},
		{"invalid json", `not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := http.Post(srv.URL+"/api/tenants", "application/json", bytes.NewBufferString(tt.body))
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
			resp.Body.Close()
		})
	}
}

func TestE2E_SeededDataIntegrity(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/tenants")
	var tenants []model.TenantPublic
	json.NewDecoder(resp.Body).Decode(&tenants)
	resp.Body.Close()
	if len(tenants) != 3 {
		t.Fatalf("expected 3 seeded tenants, got %d", len(tenants))
	}

	resp, _ = http.Get(srv.URL + "/api/transactions")
	var txns []*model.Transaction
	json.NewDecoder(resp.Body).Decode(&txns)
	resp.Body.Close()
	if len(txns) != 12 {
		t.Fatalf("expected 12 seeded transactions, got %d", len(txns))
	}

	for _, id := range []string{"t_nexus", "t_pulse", "t_orbit"} {
		resp, _ = http.Get(srv.URL + "/api/tenants/" + id + "/wallet")
		var w model.Wallet
		json.NewDecoder(resp.Body).Decode(&w)
		resp.Body.Close()
		if w.Balance != 500000000 {
			t.Errorf("expected seeded balance 500000000 for %s, got %d", id, w.Balance)
		}
	}
}

func TestE2E_GetTenantDoesNotExposeAPIKey(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/tenants/t_nexus")
	var raw map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&raw)
	resp.Body.Close()

	if _, ok := raw["api_key"]; ok {
		t.Error("GET /tenants/{id} should NOT expose api_key")
	}
}

func TestE2E_MultipleDepositsAndWithdrawals(t *testing.T) {
	srv, _ := setupE2EServer()
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/tenants", "application/json",
		bytes.NewBufferString(`{"name":"Multi Corp","brand":"MultiPay"}`))
	var createResp struct {
		ID     string `json:"id"`
		APIKey string `json:"api_key"`
	}
	json.NewDecoder(resp.Body).Decode(&createResp)
	resp.Body.Close()

	for i := 0; i < 5; i++ {
		resp, _ = postWithAuth(srv.URL+"/api/tenants/"+createResp.ID+"/deposit", `{"amount":100000}`, createResp.APIKey)
		resp.Body.Close()
	}

	for i := 0; i < 3; i++ {
		resp, _ = postWithAuth(srv.URL+"/api/tenants/"+createResp.ID+"/withdraw", `{"amount":50000}`, createResp.APIKey)
		resp.Body.Close()
	}

	resp, _ = http.Get(srv.URL + "/api/tenants/" + createResp.ID + "/wallet")
	var wallet model.Wallet
	json.NewDecoder(resp.Body).Decode(&wallet)
	resp.Body.Close()
	if wallet.Balance != 350000 {
		t.Errorf("expected balance 350000, got %d", wallet.Balance)
	}

	resp, _ = http.Get(srv.URL + "/api/tenants/" + createResp.ID + "/transactions")
	var txns []*model.Transaction
	json.NewDecoder(resp.Body).Decode(&txns)
	resp.Body.Close()
	if len(txns) != 8 {
		t.Errorf("expected 8 transactions, got %d", len(txns))
	}
}
