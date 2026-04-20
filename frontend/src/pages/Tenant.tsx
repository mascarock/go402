import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, Tenant as TenantType, Wallet, Transaction } from '../api/client'
import WalletCard from '../components/WalletCard'
import TransactionList from '../components/TransactionList'

export default function Tenant() {
  const { id } = useParams<{ id: string }>()
  const [tenant, setTenant] = useState<TenantType | null>(null)
  const [wallet, setWallet] = useState<Wallet | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [amount, setAmount] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const load = () => {
    if (!id) return
    api.getTenant(id).then(setTenant)
    api.getWallet(id).then(setWallet)
    api.getTransactions(id).then(setTransactions)
  }

  useEffect(load, [id])

  const handleDeposit = async () => {
    if (!id || !amount || !apiKey) return
    setLoading(true)
    setError('')
    try {
      await api.deposit(id, parseInt(amount) * 10000, apiKey)
      setAmount('')
      load()
    } catch (e: any) {
      setError(e.error || 'Deposit failed')
    }
    setLoading(false)
  }

  const handleWithdraw = async () => {
    if (!id || !amount || !apiKey) return
    setLoading(true)
    setError('')
    try {
      await api.withdraw(id, parseInt(amount) * 10000, apiKey)
      setAmount('')
      load()
    } catch (e: any) {
      setError(e.error || 'Withdrawal failed')
    }
    setLoading(false)
  }

  if (!tenant || !wallet) {
    return <div className="text-fg-4 text-xs font-mono py-10 text-center">Loading...</div>
  }

  return (
    <div>
      <div className="mb-8 sm:mb-10">
        <Link to="/" className="text-[10px] font-mono text-fg-4 hover:text-fg-3 transition-colors uppercase tracking-widest">
          &larr; Dashboard
        </Link>
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight mt-3 text-fg">{tenant.name}</h1>
        <p className="text-xs font-mono text-fg-3 mt-1">{tenant.brand}</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 sm:gap-8 mb-8 sm:mb-10">
        <WalletCard wallet={wallet} tenant={tenant} />

        <div className="border border-border p-5 sm:p-6">
          <h3 className="text-xs font-medium uppercase tracking-wider text-fg-3 mb-5">Actions</h3>

          {error && (
            <div className="border border-border-2 bg-surface-2 px-3 py-2.5 mb-4 text-xs font-mono text-fg-3">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label className="text-[10px] font-mono text-fg-4 uppercase tracking-widest block mb-1.5">API Key</label>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk_..."
                className="w-full px-3 py-2 bg-surface border border-border text-fg-2 font-mono text-sm focus:outline-none focus:border-border-2 placeholder:text-fg-4"
              />
            </div>
            <div>
              <label className="text-[10px] font-mono text-fg-4 uppercase tracking-widest block mb-1.5">Amount (cents)</label>
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="0"
                className="w-full px-3 py-2 bg-surface border border-border text-fg-2 font-mono text-sm focus:outline-none focus:border-border-2 placeholder:text-fg-4"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleDeposit}
                disabled={loading || !amount || !apiKey}
                className="flex-1 px-4 py-2 bg-surface-3 text-fg-2 border border-border text-xs font-medium uppercase tracking-wider hover:bg-border hover:text-fg transition-colors disabled:opacity-20 disabled:cursor-not-allowed"
              >
                Deposit
              </button>
              <button
                onClick={handleWithdraw}
                disabled={loading || !amount || !apiKey}
                className="flex-1 px-4 py-2 bg-bg text-fg-3 border border-border text-xs font-medium uppercase tracking-wider hover:bg-surface-2 hover:text-fg-2 transition-colors disabled:opacity-20 disabled:cursor-not-allowed"
              >
                Withdraw
              </button>
            </div>
          </div>
        </div>
      </div>

      <div>
        <h2 className="text-sm font-medium mb-3 tracking-tight text-fg-2">Transactions</h2>
        <TransactionList transactions={transactions} />
      </div>
    </div>
  )
}
