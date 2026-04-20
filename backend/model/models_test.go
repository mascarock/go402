package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTenantJSONExcludesAPIKey(t *testing.T) {
	original := &Tenant{
		ID:        "t_test",
		Name:      "Test Corp",
		Brand:     "TestPay",
		APIKey:    "sk_abc123",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if _, ok := raw["api_key"]; ok {
		t.Error("api_key must not appear in Tenant JSON (json:\"-\")")
	}
	if raw["id"] != "t_test" {
		t.Errorf("expected id t_test, got %v", raw["id"])
	}
}

func TestTenantPublicJSON(t *testing.T) {
	tenant := &Tenant{
		ID:        "t_test",
		Name:      "Test Corp",
		Brand:     "TestPay",
		APIKey:    "sk_secret",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	pub := tenant.Public()
	data, err := json.Marshal(pub)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if _, ok := raw["api_key"]; ok {
		t.Error("TenantPublic must not contain api_key")
	}
	if raw["id"] != "t_test" || raw["name"] != "Test Corp" || raw["brand"] != "TestPay" {
		t.Errorf("unexpected fields: %v", raw)
	}
}

func TestWalletJSONFields(t *testing.T) {
	w := &Wallet{
		TenantID:  "t_test",
		Balance:   500000000,
		Currency:  "USD",
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	expectedFields := []string{"tenant_id", "balance", "currency", "updated_at"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			t.Errorf("expected JSON field %q to be present", f)
		}
	}
}

func TestTransactionTypeConstants(t *testing.T) {
	tests := []struct {
		txType   TransactionType
		expected string
	}{
		{TxDeposit, "deposit"},
		{TxWithdrawal, "withdrawal"},
		{TxPayment, "payment"},
		{TxSettlement, "settlement"},
	}

	for _, tt := range tests {
		if string(tt.txType) != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.txType)
		}
	}
}

func TestTransactionStatusConstants(t *testing.T) {
	tests := []struct {
		status   TransactionStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.status)
		}
	}
}

func TestTransactionJSONWithMetadata(t *testing.T) {
	tx := &Transaction{
		ID:       "tx_123",
		TenantID: "t_test",
		Type:     TxPayment,
		Amount:   100000,
		Status:   StatusCompleted,
		Metadata: map[string]string{"order_id": "ord_abc"},
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Transaction
	json.Unmarshal(data, &decoded)

	if decoded.Metadata["order_id"] != "ord_abc" {
		t.Errorf("expected metadata order_id 'ord_abc', got %q", decoded.Metadata["order_id"])
	}
}

func TestTransactionJSONWithoutMetadata(t *testing.T) {
	tx := &Transaction{
		ID:       "tx_123",
		TenantID: "t_test",
		Type:     TxDeposit,
		Amount:   100000,
		Status:   StatusCompleted,
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	if _, ok := m["metadata"]; ok {
		t.Error("expected metadata to be omitted when nil")
	}
}

func TestSettlementJSON(t *testing.T) {
	s := &Settlement{
		ID:         "stl_123",
		FromTenant: "t_from",
		ToTenant:   "t_to",
		Amount:     500000,
		Status:     StatusCompleted,
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	if m["from_tenant"] != "t_from" {
		t.Errorf("expected from_tenant 't_from', got %v", m["from_tenant"])
	}
	if m["to_tenant"] != "t_to" {
		t.Errorf("expected to_tenant 't_to', got %v", m["to_tenant"])
	}
}

func TestPaymentReceiptJSON(t *testing.T) {
	r := &PaymentReceipt{
		Token:    "rcpt_abc",
		TenantID: "t_test",
		Amount:   100000,
		IssuedAt: time.Now(),
		Used:     false,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded PaymentReceipt
	json.Unmarshal(data, &decoded)

	if decoded.Token != "rcpt_abc" {
		t.Errorf("expected token 'rcpt_abc', got %q", decoded.Token)
	}
	if decoded.Used {
		t.Error("expected Used to be false")
	}
}

func TestPlatformStatsJSON(t *testing.T) {
	stats := &PlatformStats{
		TotalTenants:      5,
		TotalTransactions: 100,
		TotalVolume:       5000000,
		TotalSettlements:  10,
		ActiveWallets:     5,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded PlatformStats
	json.Unmarshal(data, &decoded)

	if decoded.TotalTenants != 5 {
		t.Errorf("expected 5 tenants, got %d", decoded.TotalTenants)
	}
	if decoded.TotalVolume != 5000000 {
		t.Errorf("expected volume 5000000, got %d", decoded.TotalVolume)
	}
}
