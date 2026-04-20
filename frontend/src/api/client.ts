const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (res.status === 402) {
    const data = await res.json()
    throw { status: 402, ...data, headers: Object.fromEntries(res.headers.entries()) }
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw err
  }
  return res.json()
}

export interface Tenant {
  id: string
  name: string
  brand: string
  created_at: string
}

export interface TenantWithKey extends Tenant {
  api_key: string
}

export interface Wallet {
  tenant_id: string
  balance: number
  currency: string
  updated_at: string
}

export interface Transaction {
  id: string
  tenant_id: string
  type: string
  amount: number
  status: string
  metadata?: Record<string, string>
  created_at: string
}

export interface Settlement {
  id: string
  from_tenant: string
  to_tenant: string
  amount: number
  status: string
  created_at: string
}

export interface PlatformStats {
  total_tenants: number
  total_transactions: number
  total_volume: number
  total_settlements: number
  active_wallets: number
}

export interface Credentials {
  tenant_id: string
  api_key: string
}

export interface PaymentResponse {
  success: boolean
  token: string
  amount: number
  message: string
}

export interface ProtectedDataResponse {
  data: string
  paid: boolean
  receipt: string
  cost: number
}

export const api = {
  getStats: () => request<PlatformStats>('/stats'),
  getTenants: () => request<Tenant[]>('/tenants'),
  getTenant: (id: string) => request<Tenant>(`/tenants/${id}`),
  getCredentials: (id: string) => request<Credentials>(`/tenants/${id}/credentials`),
  createTenant: (name: string, brand: string) =>
    request<TenantWithKey>('/tenants', { method: 'POST', body: JSON.stringify({ name, brand }) }),
  getWallet: (id: string) => request<Wallet>(`/tenants/${id}/wallet`),
  deposit: (id: string, amount: number, apiKey: string) =>
    request<Transaction>(`/tenants/${id}/deposit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: apiKey },
      body: JSON.stringify({ amount }),
    }),
  withdraw: (id: string, amount: number, apiKey: string) =>
    request<Transaction>(`/tenants/${id}/withdraw`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: apiKey },
      body: JSON.stringify({ amount }),
    }),
  getTransactions: (id: string) => request<Transaction[]>(`/tenants/${id}/transactions`),
  getAllTransactions: () => request<Transaction[]>('/transactions'),
  getSettlements: () => request<Settlement[]>('/settlements'),
  createSettlement: (from_tenant: string, to_tenant: string, amount: number, api_key: string) =>
    request<Settlement>('/settlements', { method: 'POST', body: JSON.stringify({ from_tenant, to_tenant, amount, api_key }) }),
  getProtectedData: (token?: string) =>
    request<ProtectedDataResponse>('/protected/data', {
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { 'X-Payment-Token': token } : {}),
      },
    }),
  processPayment: (tenant_id: string, api_key: string, amount: number) =>
    request<PaymentResponse>('/payments/process', {
      method: 'POST',
      body: JSON.stringify({ tenant_id, api_key, amount }),
    }),
}
