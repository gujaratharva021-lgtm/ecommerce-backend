import { useEffect, useState } from 'react'
import { getGSTSummary } from '../api/finance'
import type { GSTSummary } from '../types/finance'

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

export default function GST() {
  const [from, setFrom] = useState(daysAgoISO(30))
  const [to, setTo] = useState(todayISO())
  const [data, setData] = useState<GSTSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getGSTSummary(from, to)
      .then((res) => {
        if (!cancelled) setData(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load GST data.')
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
          <h1 className="text-lg font-semibold">GST</h1>
          <p className="text-sm text-slate-500">Output GST collected on sales for the selected period.</p>
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

      {isLoading && <p className="text-sm text-slate-500">Loading GST data...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && data && (
        <>
          <div className="border border-slate-800 bg-slate-900/40 text-slate-400 text-xs rounded-lg px-4 py-3 mb-6">
            This is output GST (tax collected on sales) only, from generated invoices. Purchase-side GST /
            input tax credit isn't tracked in the system yet, so it isn't shown here.
          </div>

          <div className="grid grid-cols-3 gap-4 mb-4">
            <SummaryCard label="Taxable Amount" value={formatCurrency(data.taxable_amount)} />
            <SummaryCard label="Total GST Collected" value={formatCurrency(data.total_gst)} tone="emerald" />
            <SummaryCard label="Invoices" value={data.invoice_count.toLocaleString('en-IN')} />
          </div>

          <div className="grid grid-cols-3 gap-4 mb-8">
            <SummaryCard label="CGST" value={formatCurrency(data.cgst_amount)} />
            <SummaryCard label="SGST" value={formatCurrency(data.sgst_amount)} />
            <SummaryCard label="IGST" value={formatCurrency(data.igst_amount)} />
          </div>

          <h2 className="text-sm font-semibold mb-3">By GST Rate</h2>
          <div className="border border-slate-800 rounded-xl overflow-hidden mb-8">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Rate</th>
                  <th className="px-4 py-2 font-medium text-right">Taxable Amount</th>
                  <th className="px-4 py-2 font-medium text-right">GST Amount</th>
                </tr>
              </thead>
              <tbody>
                {data.by_rate.length === 0 && (
                  <tr>
                    <td colSpan={3} className="px-4 py-6 text-center text-slate-500">
                      No data for this period.
                    </td>
                  </tr>
                )}
                {data.by_rate.map((row) => (
                  <tr key={row.gst_percent} className="border-t border-slate-800">
                    <td className="px-4 py-2">{row.gst_percent}%</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.taxable_amount)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.gst_amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <h2 className="text-sm font-semibold mb-3">HSN Summary</h2>
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">HSN Code</th>
                  <th className="px-4 py-2 font-medium text-right">Quantity</th>
                  <th className="px-4 py-2 font-medium text-right">Taxable Amount</th>
                  <th className="px-4 py-2 font-medium text-right">GST Amount</th>
                </tr>
              </thead>
              <tbody>
                {data.by_hsn.length === 0 && (
                  <tr>
                    <td colSpan={4} className="px-4 py-6 text-center text-slate-500">
                      No data for this period.
                    </td>
                  </tr>
                )}
                {data.by_hsn.map((row) => (
                  <tr key={row.hsn_code} className="border-t border-slate-800">
                    <td className={`px-4 py-2 ${row.hsn_code === 'Not set' ? 'text-slate-500 italic' : ''}`}>
                      {row.hsn_code}
                    </td>
                    <td className="px-4 py-2 text-right">{row.quantity}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.taxable_amount)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.gst_amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}

function SummaryCard({ label, value, tone }: { label: string; value: string; tone?: 'emerald' }) {
  return (
    <div className="border border-slate-800 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className={`text-lg font-semibold ${tone === 'emerald' ? 'text-emerald-400' : ''}`}>{value}</p>
    </div>
  )
}
