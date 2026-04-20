import { Wallet, Tenant } from '../api/client'

function formatMicrocents(mc: number): string {
  return '$' + (mc / 10000000).toFixed(2)
}

export default function WalletCard({ wallet, tenant }: { wallet: Wallet; tenant: Tenant }) {
  return (
    <div className="border border-border p-5 sm:p-6">
      <div className="flex justify-between items-start">
        <div>
          <div className="text-[10px] font-mono text-fg-4 uppercase tracking-widest">{tenant.brand}</div>
          <div className="text-3xl sm:text-4xl font-semibold mt-2 font-mono tabular-nums tracking-tight text-fg">
            {formatMicrocents(wallet.balance)}
          </div>
          <div className="text-xs font-mono text-fg-4 mt-1">{wallet.balance.toLocaleString()} microcents</div>
        </div>
        <span className="text-[10px] font-mono text-fg-4 border border-border px-1.5 py-0.5">{wallet.currency}</span>
      </div>
      <div className="mt-4 pt-4 border-t border-border">
        <span className="text-[10px] font-mono text-fg-4">{tenant.id}</span>
      </div>
    </div>
  )
}
