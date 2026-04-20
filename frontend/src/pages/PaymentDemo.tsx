import { useState, useEffect } from 'react'
import { api, Tenant } from '../api/client'

type Step = 'idle' | 'requesting' | 'payment_required' | 'paying' | 'paid' | 'accessing' | 'complete'

export default function PaymentDemo() {
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [selectedTenant, setSelectedTenant] = useState<string>('')
  const [apiKey, setApiKey] = useState('')
  const [step, setStep] = useState<Step>('idle')
  const [paymentInfo, setPaymentInfo] = useState<any>(null)
  const [token, setToken] = useState<string>('')
  const [result, setResult] = useState<any>(null)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    api.getTenants().then((t) => {
      setTenants(t)
      if (t.length > 0) setSelectedTenant(t[0].id)
    })
  }, [])

  const reset = () => {
    setStep('idle')
    setPaymentInfo(null)
    setToken('')
    setResult(null)
    setError('')
  }

  const stepRequest = async () => {
    setStep('requesting')
    setError('')
    try {
      await api.getProtectedData()
    } catch (e: any) {
      if (e.status === 402) {
        setPaymentInfo(e)
        setStep('payment_required')
      } else {
        setError(e.error || 'Unexpected error')
        setStep('idle')
      }
    }
  }

  const stepPay = async () => {
    if (!apiKey) {
      setError('Enter the tenant API key to submit payment')
      return
    }
    setStep('paying')
    try {
      const res = await api.processPayment(selectedTenant, apiKey, paymentInfo.amount)
      setToken(res.token)
      setStep('paid')
    } catch (e: any) {
      setError(e.error || 'Payment failed')
      setStep('payment_required')
    }
  }

  const stepAccess = async () => {
    setStep('accessing')
    try {
      const res = await api.getProtectedData(token)
      setResult(res)
      setStep('complete')
    } catch (e: any) {
      setError(e.error || 'Access denied')
      setStep('paid')
    }
  }

  const steps = [
    { id: 'requesting', label: 'Request Resource', desc: 'GET /api/protected/data' },
    { id: 'payment_required', label: '402 Payment Required', desc: 'Server demands micropayment' },
    { id: 'paying', label: 'Submit Payment', desc: 'POST /api/payments/process' },
    { id: 'paid', label: 'Receive Token', desc: 'Payment receipt issued' },
    { id: 'accessing', label: 'Retry with Token', desc: 'X-Payment-Token header' },
    { id: 'complete', label: 'Data Received', desc: 'Premium content delivered' },
  ]

  const currentIdx = steps.findIndex((s) => s.id === step)

  return (
    <div>
      <div className="mb-8 sm:mb-10">
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight text-fg">x402 Payment Demo</h1>
        <p className="text-xs font-mono text-fg-3 mt-1">HTTP 402 micropayment flow simulation</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8">
        <div className="lg:col-span-1 space-y-6">
          <div className="border border-border p-5">
            <h3 className="text-xs font-medium uppercase tracking-wider text-fg-3 mb-4">Configuration</h3>
            <div className="space-y-4">
              <div>
                <label className="text-[10px] font-mono text-fg-4 uppercase tracking-widest block mb-1.5">Tenant</label>
                <select
                  value={selectedTenant}
                  onChange={(e) => setSelectedTenant(e.target.value)}
                  className="w-full px-3 py-2 bg-surface border border-border text-fg-2 text-sm focus:outline-none focus:border-border-2 appearance-none"
                >
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>

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
            </div>

            <div className="mt-5">
              {step === 'idle' && (
                <button onClick={stepRequest} className="w-full px-4 py-2.5 bg-surface-3 text-fg-2 border border-border text-xs font-medium uppercase tracking-wider hover:bg-border hover:text-fg transition-colors">
                  Start Flow
                </button>
              )}
              {step === 'payment_required' && (
                <button onClick={stepPay} className="w-full px-4 py-2.5 bg-surface-3 text-fg-2 border border-border text-xs font-medium uppercase tracking-wider hover:bg-border hover:text-fg transition-colors">
                  Submit Payment
                </button>
              )}
              {step === 'paid' && (
                <button onClick={stepAccess} className="w-full px-4 py-2.5 bg-surface-3 text-fg-2 border border-border text-xs font-medium uppercase tracking-wider hover:bg-border hover:text-fg transition-colors">
                  Access Resource
                </button>
              )}
              {step === 'complete' && (
                <button onClick={reset} className="w-full px-4 py-2.5 bg-bg text-fg-3 border border-border text-xs font-medium uppercase tracking-wider hover:bg-surface-2 hover:text-fg-2 transition-colors">
                  Reset
                </button>
              )}
            </div>
          </div>

          <div className="border border-border p-5">
            <h3 className="text-xs font-medium uppercase tracking-wider text-fg-3 mb-4">Flow</h3>
            <div className="space-y-0">
              {steps.map((s, i) => (
                <div
                  key={s.id}
                  className={`flex items-start gap-3 py-2 ${
                    i <= currentIdx ? 'opacity-100' : 'opacity-20'
                  }`}
                >
                  <div className={`w-5 h-5 flex items-center justify-center text-[10px] font-mono shrink-0 mt-0.5 border ${
                    i < currentIdx ? 'border-fg-3 bg-surface-3 text-fg-2' :
                    i === currentIdx ? 'border-fg-3 text-fg-2' :
                    'border-border text-fg-4'
                  }`}>
                    {i < currentIdx ? '\u2713' : i + 1}
                  </div>
                  <div>
                    <div className="text-xs font-medium text-fg-2">{s.label}</div>
                    <div className="text-[10px] font-mono text-fg-4">{s.desc}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="lg:col-span-2">
          <div className="border border-border p-5 min-h-[300px] sm:min-h-[500px]">
            <h3 className="text-xs font-medium uppercase tracking-wider text-fg-3 mb-4">Response</h3>

            {error && (
              <div className="border border-border-2 bg-surface-2 px-3 py-2.5 mb-4 text-xs font-mono text-fg-3">
                {error}
              </div>
            )}

            {step === 'idle' && (
              <div className="text-fg-4 text-center py-20">
                <div className="text-xs font-mono">Ready to start x402 flow</div>
              </div>
            )}

            {step === 'requesting' && (
              <div className="text-fg-3 font-mono text-xs">
                <div className="animate-pulse">&gt; GET /api/protected/data</div>
              </div>
            )}

            {step === 'payment_required' && paymentInfo && (
              <div>
                <div className="text-[10px] font-mono text-fg-4 uppercase tracking-widest mb-3">HTTP 402</div>
                <pre className="bg-surface text-xs font-mono text-fg-3 p-4 border border-border overflow-x-auto">
{JSON.stringify({
  status: 402,
  headers: {
    'X-Payment-Amount': '100000',
    'X-Payment-Currency': 'USD',
    'X-Payment-Network': 'gopayments-sim',
  },
  body: {
    error: paymentInfo.error,
    amount: paymentInfo.amount,
    message: paymentInfo.message,
  }
}, null, 2)}
                </pre>
              </div>
            )}

            {step === 'paying' && (
              <div className="text-fg-3 font-mono text-xs">
                <div className="animate-pulse">&gt; POST /api/payments/process</div>
              </div>
            )}

            {step === 'paid' && token && (
              <div>
                <div className="text-[10px] font-mono text-fg-4 uppercase tracking-widest mb-3">HTTP 200</div>
                <pre className="bg-surface text-xs font-mono text-fg-3 p-4 border border-border overflow-x-auto">
{JSON.stringify({
  success: true,
  token: token,
  amount: 100000,
  message: 'Use token in X-Payment-Token header.',
}, null, 2)}
                </pre>
              </div>
            )}

            {step === 'accessing' && (
              <div className="text-fg-3 font-mono text-xs space-y-1">
                <div className="animate-pulse">&gt; GET /api/protected/data</div>
                <div className="text-fg-4 break-all">X-Payment-Token: {token}</div>
              </div>
            )}

            {step === 'complete' && result && (
              <div>
                <div className="text-[10px] font-mono text-fg-4 uppercase tracking-widest mb-3">HTTP 200 &mdash; Complete</div>
                <pre className="bg-surface text-xs font-mono text-fg-3 p-4 border border-border overflow-x-auto">
{JSON.stringify(result, null, 2)}
                </pre>
                <div className="mt-4 pt-4 border-t border-border">
                  <div className="text-[10px] font-mono text-fg-4 uppercase tracking-widest">
                    Micropayment of $0.01 deducted from tenant wallet
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
