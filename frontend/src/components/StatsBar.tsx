import { PlatformStats } from '../api/client'

function formatMicrocents(mc: number): string {
  return '$' + (mc / 10000000).toFixed(2)
}

export default function StatsBar({ stats }: { stats: PlatformStats | null }) {
  if (!stats) return null

  const items = [
    { label: 'Tenants', value: stats.total_tenants },
    { label: 'Transactions', value: stats.total_transactions },
    { label: 'Volume', value: formatMicrocents(stats.total_volume) },
    { label: 'Settlements', value: stats.total_settlements },
    { label: 'Wallets', value: stats.active_wallets },
  ]

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-px bg-n-4 mb-8 sm:mb-10 border border-n-4">
      {items.map((item) => (
        <div key={item.label} className="bg-n-0 p-3 sm:p-4">
          <div className="text-[10px] font-mono text-n-6 uppercase tracking-widest">{item.label}</div>
          <div className="text-lg sm:text-xl font-semibold mt-1 font-mono tabular-nums text-n-9">{item.value}</div>
        </div>
      ))}
    </div>
  )
}
