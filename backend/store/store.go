package store

import "github.com/mascarock/gopayments/model"

type Store interface {
	// Tenants
	ListTenants() []*model.Tenant
	GetTenant(id string) (*model.Tenant, bool)
	GetTenantByAPIKey(key string) (*model.Tenant, bool)
	CreateTenant(name, brand string) *model.Tenant

	// Wallets
	GetWallet(tenantID string) (*model.Wallet, bool)
	Deposit(tenantID string, amount int64) (*model.Transaction, error)
	Withdraw(tenantID string, amount int64) (*model.Transaction, error)

	// Transactions
	GetTransactions(tenantID string) []*model.Transaction
	GetAllTransactions() []*model.Transaction

	// Payments
	ProcessPayment(tenantID string, amount int64, metadata map[string]string) (*model.Transaction, error)
	CreateReceipt(tenantID string, amount int64) *model.PaymentReceipt
	ValidateReceipt(token string) (*model.PaymentReceipt, bool)

	// Settlements
	CreateSettlement(fromTenant, toTenant string, amount int64) (*model.Settlement, error)
	ListSettlements() []*model.Settlement

	// Stats
	GetStats() *model.PlatformStats
}
