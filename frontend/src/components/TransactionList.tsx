import { Transaction } from '../api/client'

function formatMicrocents(mc: number): string {
  const val = mc / 10000000
  return (mc >= 0 ? '+' : '-') + '$' + Math.abs(val).toFixed(4)
}

export default function TransactionList({ transactions }: { transactions: Transaction[] }) {
  if (transactions.length === 0) {
    return <div className="text-fg-4 text-center py-10 text-xs font-mono">No transactions</div>
  }

  const sorted = [...transactions].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div className="border border-border divide-y divide-border">
      {sorted.slice(0, 20).map((tx) => (
        <div key={tx.id} className="flex flex-col sm:flex-row sm:items-center sm:justify-between px-3 sm:px-4 py-2.5 gap-1 sm:gap-0">
          <div className="flex items-center gap-3 min-w-0">
            <span className="text-[10px] font-mono text-fg-3 uppercase tracking-wider w-20 shrink-0">
              {tx.type}
            </span>
            <span className="text-[10px] font-mono text-fg-4 truncate">{tx.id}</span>
          </div>
          <div className="flex items-center justify-between sm:justify-end gap-4">
            <span className="text-[10px] font-mono text-fg-4">
              {new Date(tx.created_at).toLocaleString()}
            </span>
            <span className={`font-mono text-xs tabular-nums font-medium ${tx.amount >= 0 ? 'text-fg-2' : 'text-fg-3'}`}>
              {formatMicrocents(tx.amount)}
            </span>
          </div>
        </div>
      ))}
    </div>
  )
}
