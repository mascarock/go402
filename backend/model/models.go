package model

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand"`
	APIKey    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type TenantPublic struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand"`
	CreatedAt time.Time `json:"created_at"`
}

func (t *Tenant) Public() TenantPublic {
	return TenantPublic{ID: t.ID, Name: t.Name, Brand: t.Brand, CreatedAt: t.CreatedAt}
}

type Wallet struct {
	TenantID  string    `json:"tenant_id"`
	Balance   int64     `json:"balance"` // microcents: 1 cent = 10000
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransactionType string

const (
	TxDeposit    TransactionType = "deposit"
	TxWithdrawal TransactionType = "withdrawal"
	TxPayment    TransactionType = "payment"
	TxSettlement TransactionType = "settlement"
)

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID        string            `json:"id"`
	TenantID  string            `json:"tenant_id"`
	Type      TransactionType   `json:"type"`
	Amount    int64             `json:"amount"`
	Status    TransactionStatus `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Settlement struct {
	ID         string            `json:"id"`
	FromTenant string            `json:"from_tenant"`
	ToTenant   string            `json:"to_tenant"`
	Amount     int64             `json:"amount"`
	Status     TransactionStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
}

type PaymentReceipt struct {
	Token    string    `json:"token"`
	TenantID string    `json:"tenant_id"`
	Amount   int64     `json:"amount"`
	IssuedAt time.Time `json:"issued_at"`
	Used     bool      `json:"used"`
}

type PlatformStats struct {
	TotalTenants      int   `json:"total_tenants"`
	TotalTransactions int   `json:"total_transactions"`
	TotalVolume       int64 `json:"total_volume"`
	TotalSettlements  int   `json:"total_settlements"`
	ActiveWallets     int   `json:"active_wallets"`
}
