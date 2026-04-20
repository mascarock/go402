package store

import (
	"encoding/json"
	"log"
	"os"

	"github.com/mascarock/gopayments/model"
)

type snapshot struct {
	Tenants      map[string]*model.Tenant      `json:"tenants"`
	Wallets      map[string]*model.Wallet      `json:"wallets"`
	Transactions []*model.Transaction          `json:"transactions"`
	Settlements  []*model.Settlement           `json:"settlements"`
	Receipts     map[string]*model.PaymentReceipt `json:"receipts"`
}

func (s *MemoryStore) load() bool {
	if s.path == "" {
		return false
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return false
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		log.Printf("persist: failed to decode snapshot at %s: %v", s.path, err)
		return false
	}
	if snap.Tenants != nil {
		s.tenants = snap.Tenants
	}
	if snap.Wallets != nil {
		s.wallets = snap.Wallets
	}
	if snap.Transactions != nil {
		s.transactions = snap.Transactions
	}
	if snap.Settlements != nil {
		s.settlements = snap.Settlements
	}
	if snap.Receipts != nil {
		s.receipts = snap.Receipts
	}
	log.Printf("persist: loaded snapshot from %s (%d tenants)", s.path, len(s.tenants))
	return true
}

// persistLocked writes a snapshot to disk. Caller MUST already hold s.mu (write or read).
func (s *MemoryStore) persistLocked() {
	if s.path == "" {
		return
	}
	snap := snapshot{
		Tenants:      s.tenants,
		Wallets:      s.wallets,
		Transactions: s.transactions,
		Settlements:  s.settlements,
		Receipts:     s.receipts,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		log.Printf("persist: encode failed: %v", err)
		return
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		log.Printf("persist: write failed: %v", err)
		return
	}
	if err := os.Rename(tmp, s.path); err != nil {
		log.Printf("persist: rename failed: %v", err)
	}
}
