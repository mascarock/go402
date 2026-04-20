import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, Tenant, Wallet, PlatformStats, Transaction } from '../api/client'
import StatsBar from '../components/StatsBar'
import TransactionList from '../components/TransactionList'

function formatMicrocents(mc: number): string {
  return '$' + (mc / 10000000).toFixed(2)
}

export default function Dashboard() {
  const [stats, setStats] = useState<PlatformStats | null>(null)
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [wallets, setWallets] = useState<Record<string, Wallet>>({})
  const [transactions, setTransactions] = useState<Transaction[]>([])

  useEffect(() => {
    api.getStats().then(setStats)
    api.getTenants().then(async (t) => {
      setTenants(t)
      const w: Record<string, Wallet> = {}
      await Promise.all(t.map(async (tenant) => {
        const wallet = await api.getWallet(tenant.id)
        w[tenant.id] = wallet
      }))
      setWallets(w)
    })
    api.getAllTransactions().then(setTransactions)
  }, [])

  return (
    <div>
      <div className="mb-8 sm:mb-10">
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight text-n-10">Dashboard</h1>
        <p className="text-xs font-mono text-n-8 mt-1">Multi-tenant micropayments overview</p>
      </div>

      <StatsBar stats={stats} />

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-px bg-n-4 border border-n-4 mb-8 sm:mb-10">
        {tenants.map((tenant) => (
          <Link
            key={tenant.id}
            to={`/tenants/${tenant.id}`}
            className="bg-n-0 p-4 sm:p-5 hover:bg-n-2 transition-colors group"
          >
            <div className="flex justify-between items-start">
              <div>
                <div className="text-sm font-medium text-n-9 group-hover:text-n-10 transition-colors">
                  {tenant.name}
                </div>
                <div className="text-[10px] font-mono text-n-6 mt-0.5">{tenant.brand}</div>
              </div>
              <span className="text-[10px] font-mono text-n-6">{tenant.id}</span>
            </div>
            {wallets[tenant.id] && (
              <div className="mt-4 pt-3 border-t border-n-4">
                <div className="text-lg font-semibold font-mono tabular-nums text-n-9">{formatMicrocents(wallets[tenant.id].balance)}</div>
                <div className="text-[10px] font-mono text-n-6 mt-0.5">available</div>
              </div>
            )}
          </Link>
        ))}
      </div>

      <div>
        <h2 className="text-sm font-medium mb-3 tracking-tight text-n-9">Recent Transactions</h2>
        <TransactionList transactions={transactions} />
      </div>
    </div>
  )
}
