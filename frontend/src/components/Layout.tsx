import { Link, useLocation } from 'react-router-dom'
import { ReactNode, useState } from 'react'
import { useTheme, Theme } from '../hooks/useTheme'

const nav = [
  { path: '/', label: 'Dashboard' },
  { path: '/demo', label: 'x402' },
  { path: '/settlements', label: 'Settlements' },
]

const themes: { id: Theme; label: string }[] = [
  { id: 'dark', label: 'Dark' },
  { id: 'system', label: 'Auto' },
  { id: 'light', label: 'Light' },
]

export default function Layout({ children }: { children: ReactNode }) {
  const { pathname } = useLocation()
  const [menuOpen, setMenuOpen] = useState(false)
  const [theme, setTheme] = useTheme()

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-n-4 sticky top-0 z-50 bg-n-0/95 backdrop-blur-sm">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-12 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5" onClick={() => setMenuOpen(false)}>
            <span className="font-mono font-medium text-sm tracking-tight text-n-10">gopayments</span>
            <span className="text-[10px] font-mono text-n-8 border border-n-4 px-1.5 py-0.5">sim</span>
          </Link>

          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="sm:hidden p-1.5 text-n-8 hover:text-n-9"
            aria-label="Toggle menu"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
              {menuOpen
                ? <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                : <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />}
            </svg>
          </button>

          <div className="hidden sm:flex items-center gap-4">
            <nav className="flex items-center gap-0.5">
              {nav.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`px-3 py-1 text-xs font-medium tracking-wide transition-colors ${
                    pathname === item.path
                      ? 'text-n-10 bg-n-3'
                      : 'text-n-8 hover:text-n-9'
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </nav>

            <div className="flex border border-n-4">
              {themes.map((t) => (
                <button
                  key={t.id}
                  onClick={() => setTheme(t.id)}
                  className={`px-2 py-1 text-[10px] font-mono uppercase tracking-wider transition-colors ${
                    theme === t.id ? 'bg-n-3 text-n-10' : 'text-n-8 hover:text-n-9'
                  }`}
                >
                  {t.label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {menuOpen && (
          <div className="sm:hidden border-t border-n-4 px-4 py-2 flex flex-col gap-2">
            <nav className="flex flex-col">
              {nav.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setMenuOpen(false)}
                  className={`px-3 py-2.5 text-xs font-medium tracking-wide transition-colors ${
                    pathname === item.path
                      ? 'text-n-10 bg-n-3'
                      : 'text-n-8 hover:text-n-9'
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
            <div className="flex border border-n-4 self-start">
              {themes.map((t) => (
                <button
                  key={t.id}
                  onClick={() => setTheme(t.id)}
                  className={`px-2.5 py-1.5 text-[10px] font-mono uppercase tracking-wider transition-colors ${
                    theme === t.id ? 'bg-n-3 text-n-10' : 'text-n-8 hover:text-n-9'
                  }`}
                >
                  {t.label}
                </button>
              ))}
            </div>
          </div>
        )}
      </header>

      <main className="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-6 sm:py-10">
        {children}
      </main>

      <footer className="border-t border-n-4 py-4 text-center">
        <span className="text-[10px] font-mono text-n-6 tracking-wider uppercase">GoPayments Simulation</span>
      </footer>
    </div>
  )
}
