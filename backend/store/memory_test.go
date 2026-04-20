package store

import (
	"sync"
	"testing"

	"github.com/mascarock/gopayments/model"
)

func newEmptyStore() *MemoryStore {
	return &MemoryStore{
		tenants:      make(map[string]*model.Tenant),
		wallets:      make(map[string]*model.Wallet),
		transactions: make([]*model.Transaction, 0),
		settlements:  make([]*model.Settlement, 0),
		receipts:     make(map[string]*model.PaymentReceipt),
	}
}

// --- Tenant tests ---

func TestCreateTenant(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")

	if tenant.Name != "Acme Corp" {
		t.Errorf("expected name 'Acme Corp', got %q", tenant.Name)
	}
	if tenant.Brand != "AcmePay" {
		t.Errorf("expected brand 'AcmePay', got %q", tenant.Brand)
	}
	if tenant.ID == "" {
		t.Error("expected tenant ID to be set")
	}
	if tenant.APIKey == "" {
		t.Error("expected API key to be generated")
	}
	if tenant.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateTenantAlsoCreatesWallet(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")

	wallet, ok := s.GetWallet(tenant.ID)
	if !ok {
		t.Fatal("expected wallet to be created for new tenant")
	}
	if wallet.Balance != 0 {
		t.Errorf("expected initial balance 0, got %d", wallet.Balance)
	}
	if wallet.Currency != "USD" {
		t.Errorf("expected currency 'USD', got %q", wallet.Currency)
	}
}

func TestListTenantsEmpty(t *testing.T) {
	s := newEmptyStore()
	tenants := s.ListTenants()
	if len(tenants) != 0 {
		t.Errorf("expected 0 tenants, got %d", len(tenants))
	}
}

func TestListTenants(t *testing.T) {
	s := newEmptyStore()
	s.CreateTenant("Tenant A", "BrandA")
	s.CreateTenant("Tenant B", "BrandB")

	tenants := s.ListTenants()
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(tenants))
	}
}

func TestGetTenantFound(t *testing.T) {
	s := newEmptyStore()
	created := s.CreateTenant("Acme Corp", "AcmePay")

	tenant, ok := s.GetTenant(created.ID)
	if !ok {
		t.Fatal("expected tenant to be found")
	}
	if tenant.Name != "Acme Corp" {
		t.Errorf("expected name 'Acme Corp', got %q", tenant.Name)
	}
}

func TestGetTenantNotFound(t *testing.T) {
	s := newEmptyStore()
	_, ok := s.GetTenant("nonexistent")
	if ok {
		t.Error("expected tenant to not be found")
	}
}

func TestGetTenantByAPIKey(t *testing.T) {
	s := newEmptyStore()
	created := s.CreateTenant("Acme Corp", "AcmePay")

	tenant, ok := s.GetTenantByAPIKey(created.APIKey)
	if !ok {
		t.Fatal("expected tenant to be found by API key")
	}
	if tenant.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, tenant.ID)
	}
}

func TestGetTenantByAPIKeyNotFound(t *testing.T) {
	s := newEmptyStore()
	_, ok := s.GetTenantByAPIKey("sk_invalid")
	if ok {
		t.Error("expected tenant to not be found with invalid API key")
	}
}

// --- Wallet tests ---

func TestGetWalletNotFound(t *testing.T) {
	s := newEmptyStore()
	_, ok := s.GetWallet("nonexistent")
	if ok {
		t.Error("expected wallet to not be found")
	}
}

// --- Deposit tests ---

func TestDepositSuccess(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")

	tx, err := s.Deposit(tenant.ID, 100000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != model.TxDeposit {
		t.Errorf("expected type %q, got %q", model.TxDeposit, tx.Type)
	}
	if tx.Amount != 100000 {
		t.Errorf("expected amount 100000, got %d", tx.Amount)
	}
	if tx.Status != model.StatusCompleted {
		t.Errorf("expected status %q, got %q", model.StatusCompleted, tx.Status)
	}

	wallet, _ := s.GetWallet(tenant.ID)
	if wallet.Balance != 100000 {
		t.Errorf("expected balance 100000 after deposit, got %d", wallet.Balance)
	}
}

func TestDepositMultiple(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")

	s.Deposit(tenant.ID, 100000)
	s.Deposit(tenant.ID, 200000)

	wallet, _ := s.GetWallet(tenant.ID)
	if wallet.Balance != 300000 {
		t.Errorf("expected balance 300000, got %d", wallet.Balance)
	}
}

func TestDepositWalletNotFound(t *testing.T) {
	s := newEmptyStore()
	_, err := s.Deposit("nonexistent", 100000)
	if err == nil {
		t.Error("expected error for nonexistent wallet")
	}
}

// --- Withdraw tests ---

func TestWithdrawSuccess(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")
	s.Deposit(tenant.ID, 500000)

	tx, err := s.Withdraw(tenant.ID, 200000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != model.TxWithdrawal {
		t.Errorf("expected type %q, got %q", model.TxWithdrawal, tx.Type)
	}
	if tx.Amount != 200000 {
		t.Errorf("expected amount 200000, got %d", tx.Amount)
	}

	wallet, _ := s.GetWallet(tenant.ID)
	if wallet.Balance != 300000 {
		t.Errorf("expected balance 300000, got %d", wallet.Balance)
	}
}

func TestWithdrawInsufficientBalance(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")

	_, err := s.Withdraw(tenant.ID, 100000)
	if err == nil {
		t.Error("expected error for insufficient balance")
	}
}

func TestWithdrawExactBalance(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")
	s.Deposit(tenant.ID, 100000)

	_, err := s.Withdraw(tenant.ID, 100000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wallet, _ := s.GetWallet(tenant.ID)
	if wallet.Balance != 0 {
		t.Errorf("expected balance 0, got %d", wallet.Balance)
	}
}

func TestWithdrawWalletNotFound(t *testing.T) {
	s := newEmptyStore()
	_, err := s.Withdraw("nonexistent", 100000)
	if err == nil {
		t.Error("expected error for nonexistent wallet")
	}
}

// --- Transaction tests ---

func TestGetTransactionsByTenant(t *testing.T) {
	s := newEmptyStore()
	tenantA := s.CreateTenant("A", "BrandA")
	tenantB := s.CreateTenant("B", "BrandB")

	s.Deposit(tenantA.ID, 100000)
	s.Deposit(tenantA.ID, 200000)
	s.Deposit(tenantB.ID, 300000)

	txnsA := s.GetTransactions(tenantA.ID)
	if len(txnsA) != 2 {
		t.Errorf("expected 2 transactions for tenant A, got %d", len(txnsA))
	}

	txnsB := s.GetTransactions(tenantB.ID)
	if len(txnsB) != 1 {
		t.Errorf("expected 1 transaction for tenant B, got %d", len(txnsB))
	}
}

func TestGetTransactionsEmpty(t *testing.T) {
	s := newEmptyStore()
	txns := s.GetTransactions("nonexistent")
	if len(txns) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(txns))
	}
}

func TestGetAllTransactions(t *testing.T) {
	s := newEmptyStore()
	tenantA := s.CreateTenant("A", "BrandA")
	tenantB := s.CreateTenant("B", "BrandB")

	s.Deposit(tenantA.ID, 100000)
	s.Deposit(tenantB.ID, 200000)

	all := s.GetAllTransactions()
	if len(all) != 2 {
		t.Errorf("expected 2 total transactions, got %d", len(all))
	}
}

// --- Payment tests ---

func TestProcessPaymentSuccess(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")
	s.Deposit(tenant.ID, 500000)

	meta := map[string]string{"order_id": "ord_123"}
	tx, err := s.ProcessPayment(tenant.ID, 100000, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != model.TxPayment {
		t.Errorf("expected type %q, got %q", model.TxPayment, tx.Type)
	}
	if tx.Metadata["order_id"] != "ord_123" {
		t.Errorf("expected metadata order_id 'ord_123', got %q", tx.Metadata["order_id"])
	}

	wallet, _ := s.GetWallet(tenant.ID)
	if wallet.Balance != 400000 {
		t.Errorf("expected balance 400000, got %d", wallet.Balance)
	}
}

func TestProcessPaymentInsufficientBalance(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme Corp", "AcmePay")

	_, err := s.ProcessPayment(tenant.ID, 100000, nil)
	if err == nil {
		t.Error("expected error for insufficient balance")
	}
}

func TestProcessPaymentWalletNotFound(t *testing.T) {
	s := newEmptyStore()
	_, err := s.ProcessPayment("nonexistent", 100000, nil)
	if err == nil {
		t.Error("expected error for nonexistent wallet")
	}
}

// --- Receipt tests ---

func TestCreateAndValidateReceipt(t *testing.T) {
	s := newEmptyStore()
	receipt := s.CreateReceipt("t_test", 50000)

	if receipt.Token == "" {
		t.Error("expected receipt token to be set")
	}
	if receipt.TenantID != "t_test" {
		t.Errorf("expected tenant ID 't_test', got %q", receipt.TenantID)
	}
	if receipt.Amount != 50000 {
		t.Errorf("expected amount 50000, got %d", receipt.Amount)
	}
	if receipt.Used {
		t.Error("expected receipt to not be used initially")
	}

	validated, ok := s.ValidateReceipt(receipt.Token)
	if !ok {
		t.Fatal("expected receipt validation to succeed")
	}
	if !validated.Used {
		t.Error("expected receipt to be marked as used after validation")
	}
}

func TestValidateReceiptAlreadyUsed(t *testing.T) {
	s := newEmptyStore()
	receipt := s.CreateReceipt("t_test", 50000)
	s.ValidateReceipt(receipt.Token)

	_, ok := s.ValidateReceipt(receipt.Token)
	if ok {
		t.Error("expected validation to fail for already-used receipt")
	}
}

func TestValidateReceiptNotFound(t *testing.T) {
	s := newEmptyStore()
	_, ok := s.ValidateReceipt("rcpt_nonexistent")
	if ok {
		t.Error("expected validation to fail for nonexistent receipt")
	}
}

// --- Settlement tests ---

func TestCreateSettlementSuccess(t *testing.T) {
	s := newEmptyStore()
	from := s.CreateTenant("From Corp", "FromPay")
	to := s.CreateTenant("To Corp", "ToPay")
	s.Deposit(from.ID, 1000000)

	settlement, err := s.CreateSettlement(from.ID, to.ID, 500000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settlement.FromTenant != from.ID {
		t.Errorf("expected from %q, got %q", from.ID, settlement.FromTenant)
	}
	if settlement.ToTenant != to.ID {
		t.Errorf("expected to %q, got %q", to.ID, settlement.ToTenant)
	}
	if settlement.Amount != 500000 {
		t.Errorf("expected amount 500000, got %d", settlement.Amount)
	}

	fromWallet, _ := s.GetWallet(from.ID)
	if fromWallet.Balance != 500000 {
		t.Errorf("expected from balance 500000, got %d", fromWallet.Balance)
	}
	toWallet, _ := s.GetWallet(to.ID)
	if toWallet.Balance != 500000 {
		t.Errorf("expected to balance 500000, got %d", toWallet.Balance)
	}
}

func TestCreateSettlementCreatesTransactions(t *testing.T) {
	s := newEmptyStore()
	from := s.CreateTenant("From Corp", "FromPay")
	to := s.CreateTenant("To Corp", "ToPay")
	s.Deposit(from.ID, 1000000)

	s.CreateSettlement(from.ID, to.ID, 500000)

	// Deposit creates 1 tx, settlement creates 2 => from has 2 txs, to has 1
	fromTxns := s.GetTransactions(from.ID)
	if len(fromTxns) != 2 {
		t.Errorf("expected 2 transactions for 'from' tenant, got %d", len(fromTxns))
	}
	toTxns := s.GetTransactions(to.ID)
	if len(toTxns) != 1 {
		t.Errorf("expected 1 transaction for 'to' tenant, got %d", len(toTxns))
	}
}

func TestCreateSettlementInsufficientBalance(t *testing.T) {
	s := newEmptyStore()
	from := s.CreateTenant("From Corp", "FromPay")
	to := s.CreateTenant("To Corp", "ToPay")

	_, err := s.CreateSettlement(from.ID, to.ID, 500000)
	if err == nil {
		t.Error("expected error for insufficient balance")
	}
}

func TestCreateSettlementSourceNotFound(t *testing.T) {
	s := newEmptyStore()
	to := s.CreateTenant("To Corp", "ToPay")

	_, err := s.CreateSettlement("nonexistent", to.ID, 500000)
	if err == nil {
		t.Error("expected error for nonexistent source wallet")
	}
}

func TestCreateSettlementDestinationNotFound(t *testing.T) {
	s := newEmptyStore()
	from := s.CreateTenant("From Corp", "FromPay")
	s.Deposit(from.ID, 1000000)

	_, err := s.CreateSettlement(from.ID, "nonexistent", 500000)
	if err == nil {
		t.Error("expected error for nonexistent destination wallet")
	}
}

func TestListSettlements(t *testing.T) {
	s := newEmptyStore()
	from := s.CreateTenant("From", "FP")
	to := s.CreateTenant("To", "TP")
	s.Deposit(from.ID, 2000000)

	s.CreateSettlement(from.ID, to.ID, 100000)
	s.CreateSettlement(from.ID, to.ID, 200000)

	settlements := s.ListSettlements()
	if len(settlements) != 2 {
		t.Errorf("expected 2 settlements, got %d", len(settlements))
	}
}

func TestListSettlementsEmpty(t *testing.T) {
	s := newEmptyStore()
	settlements := s.ListSettlements()
	if len(settlements) != 0 {
		t.Errorf("expected 0 settlements, got %d", len(settlements))
	}
}

// --- Stats tests ---

func TestGetStatsEmpty(t *testing.T) {
	s := newEmptyStore()
	stats := s.GetStats()
	if stats.TotalTenants != 0 {
		t.Errorf("expected 0 tenants, got %d", stats.TotalTenants)
	}
	if stats.TotalTransactions != 0 {
		t.Errorf("expected 0 transactions, got %d", stats.TotalTransactions)
	}
	if stats.TotalVolume != 0 {
		t.Errorf("expected 0 volume, got %d", stats.TotalVolume)
	}
}

func TestGetStats(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Acme", "AP")
	s.Deposit(tenant.ID, 500000)
	s.Deposit(tenant.ID, 300000)

	stats := s.GetStats()
	if stats.TotalTenants != 1 {
		t.Errorf("expected 1 tenant, got %d", stats.TotalTenants)
	}
	if stats.TotalTransactions != 2 {
		t.Errorf("expected 2 transactions, got %d", stats.TotalTransactions)
	}
	if stats.TotalVolume != 800000 {
		t.Errorf("expected volume 800000, got %d", stats.TotalVolume)
	}
	if stats.ActiveWallets != 1 {
		t.Errorf("expected 1 active wallet, got %d", stats.ActiveWallets)
	}
}

func TestGetStatsSettlementNegativeAmountExcluded(t *testing.T) {
	s := newEmptyStore()
	from := s.CreateTenant("From", "FP")
	to := s.CreateTenant("To", "TP")
	s.Deposit(from.ID, 1000000)
	s.CreateSettlement(from.ID, to.ID, 500000)

	stats := s.GetStats()
	// Volume should only include positive amounts
	// Deposit 1M + settlement debit (-500K excluded) + settlement credit (500K) = 1.5M
	if stats.TotalVolume != 1500000 {
		t.Errorf("expected volume 1500000, got %d", stats.TotalVolume)
	}
	if stats.TotalSettlements != 1 {
		t.Errorf("expected 1 settlement, got %d", stats.TotalSettlements)
	}
}

// --- Seed tests ---

func TestNewStoreHasSeededData(t *testing.T) {
	s := New()
	tenants := s.ListTenants()
	if len(tenants) != 3 {
		t.Errorf("expected 3 seeded tenants, got %d", len(tenants))
	}

	for _, tenant := range tenants {
		wallet, ok := s.GetWallet(tenant.ID)
		if !ok {
			t.Errorf("expected wallet for seeded tenant %q", tenant.ID)
		}
		if wallet.Balance != 500000000 {
			t.Errorf("expected seeded balance 500000000, got %d", wallet.Balance)
		}
	}

	allTx := s.GetAllTransactions()
	if len(allTx) != 12 {
		t.Errorf("expected 12 seeded transactions, got %d", len(allTx))
	}
}

// --- Concurrency tests ---

func TestConcurrentDeposits(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Concurrent Corp", "CC")

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			s.Deposit(tenant.ID, 10000)
		}()
	}
	wg.Wait()

	wallet, _ := s.GetWallet(tenant.ID)
	expected := int64(n * 10000)
	if wallet.Balance != expected {
		t.Errorf("expected balance %d after %d concurrent deposits, got %d", expected, n, wallet.Balance)
	}
}

func TestConcurrentDepositAndWithdraw(t *testing.T) {
	s := newEmptyStore()
	tenant := s.CreateTenant("Concurrent Corp", "CC")
	s.Deposit(tenant.ID, 10000000)

	var wg sync.WaitGroup
	n := 50
	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			s.Deposit(tenant.ID, 10000)
		}()
		go func() {
			defer wg.Done()
			s.Withdraw(tenant.ID, 10000)
		}()
	}
	wg.Wait()

	wallet, _ := s.GetWallet(tenant.ID)
	if wallet.Balance != 10000000 {
		t.Errorf("expected balance 10000000 after equal deposits/withdrawals, got %d", wallet.Balance)
	}
}

// --- ID generation tests ---

func TestGenerateIDHasPrefix(t *testing.T) {
	id := GenerateID("tx_")
	if len(id) < 4 {
		t.Errorf("expected ID with prefix, got %q", id)
	}
	if id[:3] != "tx_" {
		t.Errorf("expected prefix 'tx_', got %q", id[:3])
	}
}

func TestGenerateIDUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID("id_")
		if seen[id] {
			t.Fatalf("generated duplicate ID: %q", id)
		}
		seen[id] = true
	}
}
