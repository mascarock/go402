import { Link, useLocation } from 'react-router-dom'
import { ReactNode, useState } from 'react'

const nav = [
  { path: '/', label: 'Dashboard' },
  { path: '/demo', label: 'x402' },
  { path: '/settlements', label: 'Settlements' },
]

export default function Layout({ children }: { children: ReactNode }) {
  const { pathname } = useLocation()
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-border sticky top-0 z-50 bg-bg/95 backdrop-blur-sm">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-12 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5" onClick={() => setMenuOpen(false)}>
            <span className="font-mono font-medium text-sm tracking-tight text-fg">gopayments</span>
            <span className="text-[10px] font-mono text-fg-3 border border-border px-1.5 py-0.5">sim</span>
          </Link>

          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="sm:hidden p-1.5 text-fg-3 hover:text-fg-2"
            aria-label="Toggle menu"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
              {menuOpen
                ? <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                : <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />}
            </svg>
          </button>

          <nav className="hidden sm:flex items-center gap-0.5">
            {nav.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`px-3 py-1 text-xs font-medium tracking-wide transition-colors ${
                  pathname === item.path
                    ? 'text-fg bg-surface-3'
                    : 'text-fg-3 hover:text-fg-2'
                }`}
              >
                {item.label}
              </Link>
            ))}
          </nav>
        </div>

        {menuOpen && (
          <nav className="sm:hidden border-t border-border px-4 py-2 flex flex-col">
            {nav.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setMenuOpen(false)}
                className={`px-3 py-2.5 text-xs font-medium tracking-wide transition-colors ${
                  pathname === item.path
                    ? 'text-fg bg-surface-3'
                    : 'text-fg-3 hover:text-fg-2'
                }`}
              >
                {item.label}
              </Link>
            ))}
          </nav>
        )}
      </header>

      <main className="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-6 sm:py-10">
        {children}
      </main>

      <footer className="border-t border-border py-4 text-center">
        <span className="text-[10px] font-mono text-fg-4 tracking-wider uppercase">GoPayments Simulation</span>
      </footer>
    </div>
  )
}
