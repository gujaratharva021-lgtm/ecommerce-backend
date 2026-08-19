import { useEffect, useState } from 'react'
import { getDailySalesReport, exportDailySalesReport } from '../api/reports'
import type { DailySalesSummary } from '../types/reports'

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

export default function Reports() {
  const [date, setDate] = useState(todayISO())
  const [summary, setSummary] = useState<DailySalesSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isExporting, setIsExporting] = useState(false)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getDailySalesReport(date)
      .then((res) => {
        if (!cancelled) setSummary(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load report.')
          setSummary(null)
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [date])

  async function handleExport() {
    setIsExporting(true)
    try {
      await exportDailySalesReport(date)
    } catch {
      setError('Could not export report.')
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Reports</h1>
          <p className="text-sm text-slate-500">Daily sales summary.</p>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <input
            type="date"
            value={date}
            max={todayISO()}
            onChange={(e) => setDate(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5"
          />
          <button
            onClick={handleExport}
            disabled={isExporting || isLoading}
            className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50"
          >
            {isExporting ? 'Exporting...' : 'Export Excel'}
          </button>
        </div>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading report...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && summary && (
        <div className="grid grid-cols-3 gap-4">
          <SummaryCard label="Total Orders" value={summary.total_orders.toLocaleString('en-IN')} />
          <SummaryCard label="Delivered" value={summary.delivered_orders.toLocaleString('en-IN')} />
          <SummaryCard label="Cancelled" value={summary.cancelled_orders.toLocaleString('en-IN')} />
          <SummaryCard label="Pending" value={summary.pending_orders.toLocaleString('en-IN')} />
          <SummaryCard label="Total Revenue" value={formatCurrency(summary.total_revenue)} />
          <SummaryCard label="Avg Order Value" value={formatCurrency(summary.avg_order_value)} />
          <SummaryCard label="COD Revenue" value={formatCurrency(summary.cod_revenue)} />
          <SummaryCard label="Online Revenue" value={formatCurrency(summary.online_revenue)} />
          <SummaryCard label="COD Orders" value={summary.cod_orders.toLocaleString('en-IN')} />
          <SummaryCard label="Online Orders" value={summary.online_orders.toLocaleString('en-IN')} />
          <SummaryCard label="Delivery Charges" value={formatCurrency(summary.total_delivery_charge)} />
          <SummaryCard label="Wallet Used" value={formatCurrency(summary.total_wallet_used)} />
        </div>
      )}
    </div>
  )
}

function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-slate-800 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className="text-lg font-semibold">{value}</p>
    </div>
  )
}
