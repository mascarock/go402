package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mascarock/gopayments/model"
)

type MemoryStore struct {
	mu           sync.RWMutex
	path         string
	tenants      map[string]*model.Tenant
	wallets      map[string]*model.Wallet
	transactions []*model.Transaction
	settlements  []*model.Settlement
	receipts     map[string]*model.PaymentReceipt
}

// New creates an in-memory store with no persistence. Seeds demo data.
func New() *MemoryStore {
	s := newEmpty("")
	s.seed()
	return s
}

// NewPersistent creates a store backed by a JSON snapshot file at path.
// If the file exists it is loaded; otherwise demo data is seeded and persisted.
func NewPersistent(path string) *MemoryStore {
	s := newEmpty(path)
	if !s.load() {
		s.seed()
		s.mu.Lock()
		s.persistLocked()
		s.mu.Unlock()
	} else {
		s.logSeededKeys()
	}
	return s
}

func newEmpty(path string) *MemoryStore {
	return &MemoryStore{
		path:         path,
		tenants:      make(map[string]*model.Tenant),
		wallets:      make(map[string]*model.Wallet),
		transactions: make([]*model.Transaction, 0),
		settlements:  make([]*model.Settlement, 0),
		receipts:     make(map[string]*model.PaymentReceipt),
	}
}

func (s *MemoryStore) logSeededKeys() {
	log.Println("=== Tenant API keys (loaded from snapshot) ===")
	for _, t := range s.tenants {
		log.Printf("  %s (%s)  %s", t.ID, t.Name, t.APIKey)
	}
	log.Println("==============================================")
}

func (s *MemoryStore) seed() {
	tenants := []struct {
		id, name, brand string
	}{
		{"t_nexus", "Nexus Gaming", "NexusPlay"},
		{"t_pulse", "Pulse Finance", "PulsePay"},
		{"t_orbit", "Orbit Markets", "OrbitBet"},
	}

	log.Println("=== Seeded tenant API keys (dev only) ===")
	for _, t := range tenants {
		key := generateKey()
		s.tenants[t.id] = &model.Tenant{
			ID:        t.id,
			Name:      t.name,
			Brand:     t.brand,
			APIKey:    key,
			CreatedAt: time.Now().Add(-72 * time.Hour),
		}
		s.wallets[t.id] = &model.Wallet{
			TenantID:  t.id,
			Balance:   500000000, // $50.00 in microcents
			Currency:  "USD",
			UpdatedAt: time.Now(),
		}
		log.Printf("  %s (%s)  %s", t.id, t.name, key)
	}
	log.Println("==========================================")

	// Seed some transactions
	types := []model.TransactionType{model.TxDeposit, model.TxPayment, model.TxPayment, model.TxDeposit}
	for i, t := range tenants {
		for j := 0; j < 4; j++ {
			s.transactions = append(s.transactions, &model.Transaction{
				ID:        fmt.Sprintf("tx_%s_%d", t.id, j),
				TenantID:  t.id,
				Type:      types[(i+j)%len(types)],
				Amount:    int64((j + 1) * 50000),
				Status:    model.StatusCompleted,
				CreatedAt: time.Now().Add(-time.Duration(j*3) * time.Hour),
			})
		}
	}
}

func generateKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "sk_" + hex.EncodeToString(b)
}

func GenerateID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

// Tenant operations

func (s *MemoryStore) ListTenants() []*model.Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Tenant, 0, len(s.tenants))
	for _, t := range s.tenants {
		result = append(result, t)
	}
	return result
}

func (s *MemoryStore) GetTenant(id string) (*model.Tenant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tenants[id]
	return t, ok
}

func (s *MemoryStore) GetTenantByAPIKey(key string) (*model.Tenant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tenants {
		if t.APIKey == key {
			return t, true
		}
	}
	return nil, false
}

func (s *MemoryStore) CreateTenant(name, brand string) *model.Tenant {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := &model.Tenant{
		ID:        GenerateID("t_"),
		Name:      name,
		Brand:     brand,
		APIKey:    generateKey(),
		CreatedAt: time.Now(),
	}
	s.tenants[t.ID] = t
	s.wallets[t.ID] = &model.Wallet{
		TenantID:  t.ID,
		Balance:   0,
		Currency:  "USD",
		UpdatedAt: time.Now(),
	}
	s.persistLocked()
	return t
}

// Wallet operations

func (s *MemoryStore) GetWallet(tenantID string) (*model.Wallet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.wallets[tenantID]
	return w, ok
}

func (s *MemoryStore) Deposit(tenantID string, amount int64) (*model.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.wallets[tenantID]
	if !ok {
		return nil, fmt.Errorf("wallet not found")
	}
	w.Balance += amount
	w.UpdatedAt = time.Now()

	tx := &model.Transaction{
		ID:        GenerateID("tx_"),
		TenantID:  tenantID,
		Type:      model.TxDeposit,
		Amount:    amount,
		Status:    model.StatusCompleted,
		CreatedAt: time.Now(),
	}
	s.transactions = append(s.transactions, tx)
	s.persistLocked()
	return tx, nil
}

func (s *MemoryStore) Withdraw(tenantID string, amount int64) (*model.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.wallets[tenantID]
	if !ok {
		return nil, fmt.Errorf("wallet not found")
	}
	if w.Balance < amount {
		return nil, fmt.Errorf("insufficient balance")
	}
	w.Balance -= amount
	w.UpdatedAt = time.Now()

	tx := &model.Transaction{
		ID:        GenerateID("tx_"),
		TenantID:  tenantID,
		Type:      model.TxWithdrawal,
		Amount:    amount,
		Status:    model.StatusCompleted,
		CreatedAt: time.Now(),
	}
	s.transactions = append(s.transactions, tx)
	s.persistLocked()
	return tx, nil
}

// Transaction operations

func (s *MemoryStore) GetTransactions(tenantID string) []*model.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Transaction, 0)
	for _, tx := range s.transactions {
		if tx.TenantID == tenantID {
			result = append(result, tx)
		}
	}
	return result
}

func (s *MemoryStore) GetAllTransactions() []*model.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Transaction, len(s.transactions))
	copy(result, s.transactions)
	return result
}

// Payment operations

func (s *MemoryStore) ProcessPayment(tenantID string, amount int64, metadata map[string]string) (*model.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.wallets[tenantID]
	if !ok {
		return nil, fmt.Errorf("wallet not found")
	}
	if w.Balance < amount {
		return nil, fmt.Errorf("insufficient balance")
	}
	w.Balance -= amount
	w.UpdatedAt = time.Now()

	tx := &model.Transaction{
		ID:        GenerateID("tx_"),
		TenantID:  tenantID,
		Type:      model.TxPayment,
		Amount:    amount,
		Status:    model.StatusCompleted,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
	s.transactions = append(s.transactions, tx)
	s.persistLocked()
	return tx, nil
}

func (s *MemoryStore) CreateReceipt(tenantID string, amount int64) *model.PaymentReceipt {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := &model.PaymentReceipt{
		Token:    GenerateID("rcpt_"),
		TenantID: tenantID,
		Amount:   amount,
		IssuedAt: time.Now(),
	}
	s.receipts[r.Token] = r
	s.persistLocked()
	return r
}

func (s *MemoryStore) ValidateReceipt(token string) (*model.PaymentReceipt, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.receipts[token]
	if !ok || r.Used {
		return nil, false
	}
	r.Used = true
	s.persistLocked()
	return r, true
}

// Settlement operations

func (s *MemoryStore) CreateSettlement(fromTenant, toTenant string, amount int64) (*model.Settlement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromWallet, ok := s.wallets[fromTenant]
	if !ok {
		return nil, fmt.Errorf("source wallet not found")
	}
	toWallet, ok := s.wallets[toTenant]
	if !ok {
		return nil, fmt.Errorf("destination wallet not found")
	}
	if fromWallet.Balance < amount {
		return nil, fmt.Errorf("insufficient balance for settlement")
	}

	fromWallet.Balance -= amount
	toWallet.Balance += amount
	fromWallet.UpdatedAt = time.Now()
	toWallet.UpdatedAt = time.Now()

	settlement := &model.Settlement{
		ID:         GenerateID("stl_"),
		FromTenant: fromTenant,
		ToTenant:   toTenant,
		Amount:     amount,
		Status:     model.StatusCompleted,
		CreatedAt:  time.Now(),
	}
	s.settlements = append(s.settlements, settlement)

	// Record transactions for both parties
	s.transactions = append(s.transactions, &model.Transaction{
		ID:       GenerateID("tx_"),
		TenantID: fromTenant,
		Type:     model.TxSettlement,
		Amount:   -amount,
		Status:   model.StatusCompleted,
		Metadata: map[string]string{"settlement_id": settlement.ID, "to": toTenant},
		CreatedAt: time.Now(),
	})
	s.transactions = append(s.transactions, &model.Transaction{
		ID:       GenerateID("tx_"),
		TenantID: toTenant,
		Type:     model.TxSettlement,
		Amount:   amount,
		Status:   model.StatusCompleted,
		Metadata: map[string]string{"settlement_id": settlement.ID, "from": fromTenant},
		CreatedAt: time.Now(),
	})

	s.persistLocked()
	return settlement, nil
}

func (s *MemoryStore) ListSettlements() []*model.Settlement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Settlement, len(s.settlements))
	copy(result, s.settlements)
	return result
}

// Stats

func (s *MemoryStore) GetStats() *model.PlatformStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var totalVolume int64
	for _, tx := range s.transactions {
		if tx.Amount > 0 {
			totalVolume += tx.Amount
		}
	}
	return &model.PlatformStats{
		TotalTenants:      len(s.tenants),
		TotalTransactions: len(s.transactions),
		TotalVolume:       totalVolume,
		TotalSettlements:  len(s.settlements),
		ActiveWallets:     len(s.wallets),
	}
}
