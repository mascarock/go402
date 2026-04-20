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
    return <div className="text-n-6 text-xs font-mono py-10 text-center">Loading...</div>
  }

  return (
    <div>
      <div className="mb-8 sm:mb-10">
        <Link to="/" className="text-[10px] font-mono text-n-6 hover:text-n-8 transition-colors uppercase tracking-widest">
          &larr; Dashboard
        </Link>
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight mt-3 text-n-10">{tenant.name}</h1>
        <p className="text-xs font-mono text-n-8 mt-1">{tenant.brand}</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 sm:gap-8 mb-8 sm:mb-10">
        <WalletCard wallet={wallet} tenant={tenant} />

        <div className="border border-n-4 p-5 sm:p-6">
          <h3 className="text-xs font-medium uppercase tracking-wider text-n-8 mb-5">Actions</h3>

          {error && (
            <div className="border border-n-5 bg-n-2 px-3 py-2.5 mb-4 text-xs font-mono text-n-8">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label className="text-[10px] font-mono text-n-6 uppercase tracking-widest block mb-1.5">API Key</label>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk_..."
                className="w-full px-3 py-2 bg-n-1 border border-n-4 text-n-9 font-mono text-sm focus:outline-none focus:border-n-5 placeholder:text-n-6"
              />
            </div>
            <div>
              <label className="text-[10px] font-mono text-n-6 uppercase tracking-widest block mb-1.5">Amount (cents)</label>
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="0"
                className="w-full px-3 py-2 bg-n-1 border border-n-4 text-n-9 font-mono text-sm focus:outline-none focus:border-n-5 placeholder:text-n-6"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleDeposit}
                disabled={loading || !amount || !apiKey}
                className="flex-1 px-4 py-2 bg-n-3 text-n-9 border border-n-4 text-xs font-medium uppercase tracking-wider hover:bg-n-4 hover:text-n-10 transition-colors disabled:opacity-20 disabled:cursor-not-allowed"
              >
                Deposit
              </button>
              <button
                onClick={handleWithdraw}
                disabled={loading || !amount || !apiKey}
                className="flex-1 px-4 py-2 bg-n-0 text-n-8 border border-n-4 text-xs font-medium uppercase tracking-wider hover:bg-n-2 hover:text-n-9 transition-colors disabled:opacity-20 disabled:cursor-not-allowed"
              >
                Withdraw
              </button>
            </div>
          </div>
        </div>
      </div>

      <div>
        <h2 className="text-sm font-medium mb-3 tracking-tight text-n-9">Transactions</h2>
        <TransactionList transactions={transactions} />
      </div>
    </div>
  )
}
