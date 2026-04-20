import { useEffect, useState } from 'react'
import { api, Tenant, Settlement } from '../api/client'

function formatMicrocents(mc: number): string {
  return '$' + (Math.abs(mc) / 10000000).toFixed(2)
}

export default function Settlements() {
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [settlements, setSettlements] = useState<Settlement[]>([])
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [amount, setAmount] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const load = () => {
    api.getTenants().then((t) => {
      setTenants(t)
      if (t.length >= 2) {
        setFrom(t[0].id)
        setTo(t[1].id)
      }
    })
    api.getSettlements().then(setSettlements)
  }

  useEffect(load, [])

  const handleCreate = async () => {
    if (!from || !to || !amount || !apiKey) return
    setLoading(true)
    setError('')
    try {
      await api.createSettlement(from, to, parseInt(amount) * 10000, apiKey)
      setAmount('')
      load()
    } catch (e: any) {
      setError(e.error || 'Settlement failed')
    }
    setLoading(false)
  }

  const getTenantName = (id: string) => tenants.find((t) => t.id === id)?.name || id

  return (
    <div>
      <div className="mb-8 sm:mb-10">
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight text-n-10">Settlements</h1>
        <p className="text-xs font-mono text-n-8 mt-1">Inter-tenant fund transfers</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8">
        <div className="lg:col-span-1">
          <div className="border border-n-4 p-5">
            <h3 className="text-xs font-medium uppercase tracking-wider text-n-8 mb-4">New Settlement</h3>

            {error && (
              <div className="border border-n-5 bg-n-2 px-3 py-2.5 mb-4 text-xs font-mono text-n-8">
                {error}
              </div>
            )}

            <div className="space-y-4">
              <div>
                <label className="text-[10px] font-mono text-n-6 uppercase tracking-widest block mb-1.5">From</label>
                <select
                  value={from}
                  onChange={(e) => setFrom(e.target.value)}
                  className="w-full px-3 py-2 bg-n-1 border border-n-4 text-n-9 text-sm focus:outline-none focus:border-n-5 appearance-none"
                >
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-[10px] font-mono text-n-6 uppercase tracking-widest block mb-1.5">API Key (sender)</label>
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="sk_..."
                  className="w-full px-3 py-2 bg-n-1 border border-n-4 text-n-9 font-mono text-sm focus:outline-none focus:border-n-5 placeholder:text-n-6"
                />
              </div>

              <div className="text-center text-n-6 text-xs font-mono py-1">&darr;</div>

              <div>
                <label className="text-[10px] font-mono text-n-6 uppercase tracking-widest block mb-1.5">To</label>
                <select
                  value={to}
                  onChange={(e) => setTo(e.target.value)}
                  className="w-full px-3 py-2 bg-n-1 border border-n-4 text-n-9 text-sm focus:outline-none focus:border-n-5 appearance-none"
                >
                  {tenants.filter((t) => t.id !== from).map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
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

              <button
                onClick={handleCreate}
                disabled={loading || !amount || !apiKey || from === to}
                className="w-full px-4 py-2.5 bg-n-3 text-n-9 border border-n-4 text-xs font-medium uppercase tracking-wider hover:bg-n-4 hover:text-n-10 transition-colors disabled:opacity-20 disabled:cursor-not-allowed"
              >
                {loading ? 'Processing...' : 'Execute'}
              </button>
            </div>
          </div>
        </div>

        <div className="lg:col-span-2">
          <div className="border border-n-4">
            <div className="px-5 py-3 border-b border-n-4">
              <h3 className="text-xs font-medium uppercase tracking-wider text-n-8">History</h3>
            </div>
            {settlements.length === 0 ? (
              <div className="text-n-6 text-center py-16 text-xs font-mono">
                No settlements yet
              </div>
            ) : (
              <div className="divide-y divide-n-4">
                {[...settlements].reverse().map((s) => (
                  <div key={s.id} className="flex flex-col sm:flex-row sm:items-center sm:justify-between px-5 py-3 gap-1 sm:gap-0">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="min-w-0">
                        <div className="text-xs font-medium text-n-9 truncate">{getTenantName(s.from_tenant)}</div>
                      </div>
                      <span className="text-n-6 shrink-0 text-xs font-mono">&rarr;</span>
                      <div className="min-w-0">
                        <div className="text-xs font-medium text-n-9 truncate">{getTenantName(s.to_tenant)}</div>
                      </div>
                    </div>
                    <div className="flex items-center justify-between sm:justify-end gap-4">
                      <span className="font-mono text-xs tabular-nums font-medium text-n-9">{formatMicrocents(s.amount)}</span>
                      <span className="text-[10px] font-mono text-n-6">{new Date(s.created_at).toLocaleString()}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
