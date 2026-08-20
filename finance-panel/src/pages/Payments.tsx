import { useEffect, useState } from 'react'
import { getPaymentReconciliation } from '../api/finance'
import type { PaymentReconciliation } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

function daysAgoISO(days: number) {
  const d = new Date()
  d.setDate(d.getDate() - days)
  return d.toISOString().slice(0, 10)
}

export default function Payments() {
  const [from, setFrom] = useState(daysAgoISO(30))
  const [to, setTo] = useState(todayISO())
  const [data, setData] = useState<PaymentReconciliation | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getPaymentReconciliation(from, to)
      .then((res) => {
        if (!cancelled) setData(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load payment data.')
          setData(null)
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [from, to])

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Payments &amp; Refunds</h1>
          <p className="text-sm text-slate-500">Collection, pending, and refund status for the selected period.</p>
        </div>

        <div className="flex items-center gap-2 text-sm">
          <input
            type="date"
            value={from}
            max={to}
            onChange={(e) => setFrom(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5"
          />
          <span className="text-slate-500">to</span>
          <input
            type="date"
            value={to}
            min={from}
            max={todayISO()}
            onChange={(e) => setTo(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5"
          />
        </div>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading payment data...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && data && (
        <>
          <div className="border border-slate-800 bg-slate-900/40 text-slate-400 text-xs rounded-lg px-4 py-3 mb-6">
            Payment instrument (UPI vs Card vs Wallet) isn't tracked separately from the gateway yet — only
            Online vs COD is split below. Add Razorpay's payment method field to the Payment record to break
            this down further.
          </div>

          <div className="grid grid-cols-3 gap-4 mb-4">
            <SummaryCard label="Collected" value={formatCurrency(data.total_collected)} tone="emerald" />
            <SummaryCard label="Pending" value={formatCurrency(data.total_pending)} tone="amber" />
            <SummaryCard label="Refunded" value={formatCurrency(data.total_refunded)} tone="red" />
          </div>

          <div className="grid grid-cols-2 gap-4 mb-8">
            <SummaryCard label="Online Collected" value={formatCurrency(data.online_collected)} />
            <SummaryCard label="COD Collected" value={formatCurrency(data.cod_collected)} />
          </div>

          <h2 className="text-sm font-semibold mb-3">Order Counts by Status</h2>
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Status</th>
                  <th className="px-4 py-2 font-medium text-right">Orders</th>
                </tr>
              </thead>
              <tbody>
                <tr className="border-t border-slate-800">
                  <td className="px-4 py-2">Paid</td>
                  <td className="px-4 py-2 text-right">{data.count_paid}</td>
                </tr>
                <tr className="border-t border-slate-800">
                  <td className="px-4 py-2">Pending</td>
                  <td className="px-4 py-2 text-right">{data.count_pending}</td>
                </tr>
                <tr className="border-t border-slate-800">
                  <td className="px-4 py-2">Failed</td>
                  <td className="px-4 py-2 text-right">{data.count_failed}</td>
                </tr>
                <tr className="border-t border-slate-800">
                  <td className="px-4 py-2">Refunded</td>
                  <td className="px-4 py-2 text-right">{data.count_refunded}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}

function SummaryCard({
  label,
  value,
  tone,
}: {
  label: string
  value: string
  tone?: 'emerald' | 'amber' | 'red'
}) {
  const toneClass =
    tone === 'emerald'
      ? 'text-emerald-400'
      : tone === 'amber'
      ? 'text-amber-400'
      : tone === 'red'
      ? 'text-red-400'
      : ''
  return (
    <div className="border border-slate-800 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className={`text-lg font-semibold ${toneClass}`}>{value}</p>
    </div>
  )
}
