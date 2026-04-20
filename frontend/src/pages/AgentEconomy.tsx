import { useState, useEffect, useRef } from 'react'
import { api } from '../api/client'

const AGENTS = [
  { id: 't_nexus', name: 'NexusBot', role: 'Orchestrator', symbol: '⬡' },
  { id: 't_pulse', name: 'PulseBot', role: 'Analyst', symbol: '◈' },
  { id: 't_orbit', name: 'OrbitBot', role: 'Trader', symbol: '◉' },
]

// 1 cent = 10_000 units → $1 = 1_000_000 units
const fmt = (units: number) => `$${(units / 1_000_000).toFixed(4)}`

type Dir = '→' | '←' | '·'
type LogEntry = { id: number; dir: Dir; agent?: string; text: string; sub?: string; ok?: boolean; err?: boolean }
type Stage = 'idle' | 'running' | 'done' | 'error'

let _id = 0

export default function AgentEconomy() {
  const [wallets, setWallets] = useState<Record<string, number>>({})
  const [log, setLog] = useState<LogEntry[]>([])
  const [stage, setStage] = useState<Stage>('idle')
  const [action, setAction] = useState('')
  const [activeId, setActiveId] = useState<string | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const load = () =>
      Promise.all(AGENTS.map((a) => api.getWallet(a.id))).then((ws) => {
        const m: Record<string, number> = {}
        ws.forEach((w, i) => (m[AGENTS[i].id] = w.balance))
        setWallets(m)
      }).catch(() => {})
    load()
    const t = setInterval(load, 2000)
    return () => clearInterval(t)
  }, [])

  const push = (dir: Dir, text: string, agent?: string, sub?: string, ok?: boolean, err?: boolean) => {
    const entry: LogEntry = { id: ++_id, dir, text, agent, sub, ok, err }
    setLog((prev) => [...prev, entry])
    setTimeout(() => bottomRef.current?.scrollIntoView({ behavior: 'smooth' }), 40)
  }

  const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

  const runDemo = async () => {
    setStage('running')
    setLog([])
    setAction('')

    try {
      // 1 · Fetch NexusBot credentials (public endpoint)
      setAction('Loading agent credentials…')
      setActiveId('t_nexus')
      push('→', 'Fetch NexusBot credentials', 'NexusBot', 'GET /api/tenants/t_nexus/credentials')
      await sleep(600)
      const creds = await api.getCredentials('t_nexus')
      push('←', 'Credentials ready', 'NexusBot', `id: ${creds.tenant_id}`, true)
      await sleep(500)

      // 2 · Top-up wallet if needed
      const wallet = await api.getWallet('t_nexus')
      if (wallet.balance < 500_000) {
        setAction('Depositing funds into NexusBot wallet…')
        push('→', 'Wallet low — depositing $1.00', 'NexusBot', 'POST /api/tenants/t_nexus/deposit')
        await sleep(500)
        await api.deposit('t_nexus', 1_000_000, creds.api_key)
        push('←', '$1.00 deposited', 'NexusBot', undefined, true)
        await sleep(500)
      }

      // 3 · Hit protected endpoint — expect 402
      setAction('NexusBot requesting premium market data…')
      push('·', 'MISSION: get premium data → pay for it → hire PulseBot to analyze it', 'NexusBot')
      await sleep(900)
      push('→', 'Request premium market data', 'NexusBot', 'GET /api/protected/data')
      await sleep(700)
      let paymentInfo: any
      try {
        await api.getProtectedData()
      } catch (e: any) {
        if (e.status === 402) {
          paymentInfo = e
          push('←', 'HTTP 402 Payment Required', 'NexusBot', `amount: ${paymentInfo.amount} units (${fmt(paymentInfo.amount)})`, false, true)
        } else throw e
      }
      await sleep(600)

      // 4 · Pay for access
      setAction('NexusBot authorising micropayment…')
      push('→', 'Submit micropayment to unlock data', 'NexusBot', 'POST /api/payments/process')
      await sleep(800)
      const payment = await api.processPayment('t_nexus', creds.api_key, paymentInfo.amount)
      push('←', 'Payment accepted — one-time token issued', 'NexusBot', `token: ${payment.token.slice(0, 20)}…`, true)
      await sleep(600)

      // 5 · Retry with token
      setAction('NexusBot retrying with payment token…')
      push('→', 'Retry request with token', 'NexusBot', 'GET /api/protected/data  +  X-Payment-Token')
      await sleep(700)
      const data = await api.getProtectedData(payment.token)
      push('←', 'Data unlocked!', 'NexusBot', data.data, true)
      await sleep(900)

      // 6 · Hire PulseBot via settlement
      setAction('NexusBot hiring PulseBot to analyse the data…')
      setActiveId('t_pulse')
      push('·', 'NexusBot → PulseBot: "Analyse this for $0.05"', 'NexusBot')
      await sleep(900)
      push('→', 'Create settlement: NexusBot pays PulseBot', 'NexusBot', 'POST /api/settlements  amount: 50 000 units ($0.05)')
      await sleep(700)
      const settlement = await api.createSettlement('t_nexus', 't_pulse', 50_000, creds.api_key)
      push('←', 'Settlement complete — PulseBot paid', 'PulseBot', `id: ${settlement.id.slice(0, 12)}…  $0.05 transferred`, true)
      await sleep(600)

      // Done
      setActiveId(null)
      setAction('Mission complete. All balances updated.')
      push('·', 'Done. Watch the wallet balances above refresh.', undefined, undefined, true)
      setStage('done')
    } catch (err: any) {
      push('←', `Error: ${err.error || err.message || 'unknown'}`, undefined, undefined, false, true)
      setStage('error')
      setAction('Something went wrong.')
      setActiveId(null)
    }
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-semibold tracking-tight text-n-10">Agent Economy</h1>
          <p className="text-xs font-mono text-n-8 mt-1">AI agents that autonomously hire and pay other agents</p>
        </div>
        <button
          onClick={runDemo}
          disabled={stage === 'running'}
          className="shrink-0 px-5 py-2.5 bg-n-9 text-n-0 text-xs font-medium uppercase tracking-wider hover:bg-n-10 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {stage === 'running' ? 'Running…' : stage === 'idle' ? 'Run Demo' : 'Run Again'}
        </button>
      </div>

      {/* Agent wallets */}
      <div className="grid grid-cols-3 gap-3 sm:gap-4 mb-6">
        {AGENTS.map((a) => (
          <div
            key={a.id}
            className={`border p-4 transition-colors ${activeId === a.id ? 'border-n-8 bg-n-2' : 'border-n-4'}`}
          >
            <div className="flex items-center gap-2 mb-3">
              <span className="text-base sm:text-lg leading-none">{a.symbol}</span>
              <div className="min-w-0">
                <div className="text-xs font-medium text-n-10 truncate">{a.name}</div>
                <div className="text-[10px] font-mono text-n-6 uppercase tracking-wider">{a.role}</div>
              </div>
              {activeId === a.id && (
                <div className="ml-auto w-1.5 h-1.5 rounded-full bg-n-9 animate-pulse shrink-0" />
              )}
            </div>
            <div className="font-mono text-sm text-n-10">
              {wallets[a.id] !== undefined ? fmt(wallets[a.id]) : '—'}
            </div>
            <div className="text-[10px] font-mono text-n-6 mt-0.5">live balance</div>
          </div>
        ))}
      </div>

      {/* Current action banner */}
      {action && (
        <div className="border border-n-4 px-4 py-3 mb-6 flex items-center gap-3">
          {stage === 'running' && <div className="w-1.5 h-1.5 rounded-full bg-n-8 animate-pulse shrink-0" />}
          <span className="text-xs font-mono text-n-8">{action}</span>
        </div>
      )}

      {/* Activity log */}
      <div className="border border-n-4">
        <div className="px-4 py-3 border-b border-n-4 flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-n-8">Activity Log</span>
          {stage === 'idle' && <span className="text-[10px] font-mono text-n-6">Press "Run Demo" to start</span>}
        </div>
        <div className="p-4 min-h-[260px] max-h-[420px] overflow-y-auto font-mono text-xs space-y-2.5">
          {log.length === 0 && (
            <div className="text-n-6 text-center py-16">Waiting to start…</div>
          )}
          {log.map((e) => (
            <div key={e.id} className="flex gap-2.5">
              <span className={`shrink-0 w-3 ${e.dir === '→' ? 'text-n-7' : e.dir === '←' ? (e.err ? 'text-n-8' : 'text-n-10') : 'text-n-5'}`}>
                {e.dir}
              </span>
              <div className="min-w-0">
                {e.agent && <span className="text-n-6 mr-1.5">[{e.agent}]</span>}
                <span className={e.ok ? 'text-n-10' : e.err ? 'text-n-8' : 'text-n-7'}>{e.text}</span>
                {e.sub && <div className="text-[10px] text-n-6 mt-0.5 break-all">{e.sub}</div>}
              </div>
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </div>

      {/* Explainer */}
      <div className="mt-6 border border-n-4 p-5">
        <h3 className="text-xs font-medium uppercase tracking-wider text-n-8 mb-4">What's happening</h3>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-5 text-[11px] font-mono text-n-7 leading-relaxed">
          <div>
            <div className="text-n-9 font-medium mb-1">1 · Pay-per-use API (x402)</div>
            NexusBot hits a locked endpoint and gets HTTP 402. It submits a micropayment, receives a one-time token, and retries — unlocking the data automatically, with no human involved.
          </div>
          <div>
            <div className="text-n-9 font-medium mb-1">2 · Agent-to-agent settlement</div>
            NexusBot pays PulseBot directly for analysis work. No invoices, no manual transfers — one API call moves money between agent wallets in real time.
          </div>
          <div>
            <div className="text-n-9 font-medium mb-1">3 · Full audit trail</div>
            Every payment is recorded as a transaction. Balances update live. You can audit every cent every agent spent — check the Dashboard and Settlements tabs.
          </div>
        </div>
      </div>
    </div>
  )
}
