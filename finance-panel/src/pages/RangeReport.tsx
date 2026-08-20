import { useEffect, useState } from 'react'
import { getRangeSalesReport, exportRangeSalesReport } from '../api/reports'
import type { RangeSalesSummary } from '../types/reports'

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

export default function RangeReport() {
  const [from, setFrom] = useState(daysAgoISO(6))
  const [to, setTo] = useState(todayISO())
  const [summary, setSummary] = useState<RangeSalesSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isExporting, setIsExporting] = useState(false)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getRangeSalesReport(from, to)
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
  }, [from, to])

  async function handleExport() {
    setIsExporting(true)
    try {
      await exportRangeSalesReport(from, to)
    } catch {
      setError('Could not export report.')
    } finally {
      setIsExporting(false)
    }
  }

  function setPreset(days: number) {
    setFrom(daysAgoISO(days - 1))
    setTo(todayISO())
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Custom Range Report</h1>
          <p className="text-sm text-slate-500">
            Sales, GST (output + vendor), HSN and orders for any date range — ready to hand to your CA.
          </p>
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
          <button
            onClick={handleExport}
            disabled={isExporting || isLoading}
            className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50"
          >
            {isExporting ? 'Exporting...' : 'Export Excel'}
          </button>
        </div>
      </div>

      <div className="flex gap-2 mb-6 text-sm">
        <button onClick={() => setPreset(7)} className="px-3 py-1.5 rounded-lg text-slate-400 hover:bg-slate-800 transition-colors">
          Last 7 Days
        </button>
        <button onClick={() => setPreset(30)} className="px-3 py-1.5 rounded-lg text-slate-400 hover:bg-slate-800 transition-colors">
          Last 30 Days
        </button>
        <button onClick={() => setPreset(90)} className="px-3 py-1.5 rounded-lg text-slate-400 hover:bg-slate-800 transition-colors">
          Last 90 Days
        </button>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading report...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && summary && (
        <>
          <h2 className="text-sm font-semibold mb-3 text-slate-300">Sales</h2>
          <div className="grid grid-cols-3 gap-4 mb-8">
            <SummaryCard label="Total Orders" value={summary.total_orders.toLocaleString('en-IN')} />
            <SummaryCard label="Delivered" value={summary.delivered_orders.toLocaleString('en-IN')} />
            <SummaryCard label="Cancelled" value={summary.cancelled_orders.toLocaleString('en-IN')} />
            <SummaryCard label="Total Revenue" value={formatCurrency(summary.total_revenue)} />
            <SummaryCard label="Avg Order Value" value={formatCurrency(summary.avg_order_value)} />
            <SummaryCard label="Delivery Charges" value={formatCurrency(summary.total_delivery_charge)} />
            <SummaryCard label="COD Revenue" value={formatCurrency(summary.cod_revenue)} />
            <SummaryCard label="Online Revenue" value={formatCurrency(summary.online_revenue)} />
            <SummaryCard label="Wallet Used" value={formatCurrency(summary.total_wallet_used)} />
          </div>

          <h2 className="text-sm font-semibold mb-3 text-slate-300">GST</h2>
          <div className="grid grid-cols-3 gap-4 mb-4">
            <SummaryCard label="Taxable Amount" value={formatCurrency(summary.taxable_amount)} />
            <SummaryCard label="Total Output GST (Sales)" value={formatCurrency(summary.total_output_gst)} tone="emerald" />
            <SummaryCard label="Total Vendor GST (Purchases)" value={formatCurrency(summary.total_vendor_gst)} tone="amber" />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <SummaryCard label="CGST" value={formatCurrency(summary.cgst_amount)} />
            <SummaryCard label="SGST" value={formatCurrency(summary.sgst_amount)} />
            <SummaryCard label="IGST" value={formatCurrency(summary.igst_amount)} />
          </div>

          <div className="border border-slate-800 bg-slate-900/40 text-slate-400 text-xs rounded-lg px-4 py-3 mt-8">
            Click "Export Excel" for the full CA-ready workbook: Summary, Orders, GST By Rate, GST By HSN, and Vendor GST sheets for this date range.
          </div>
        </>
      )}
    </div>
  )
}

function SummaryCard({ label, value, tone }: { label: string; value: string; tone?: 'emerald' | 'amber' }) {
  const toneClass = tone === 'emerald' ? 'text-emerald-400' : tone === 'amber' ? 'text-amber-400' : ''
  return (
    <div className="border border-slate-800 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className={`text-lg font-semibold ${toneClass}`}>{value}</p>
    </div>
  )
}
