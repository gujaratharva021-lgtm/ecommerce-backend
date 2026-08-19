import { useEffect, useState } from 'react'
import { getRevenue } from '../api/finance'
import type { RevenueResponse } from '../types/finance'

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

export default function Revenue() {
  const [from, setFrom] = useState(daysAgoISO(30))
  const [to, setTo] = useState(todayISO())
  const [data, setData] = useState<RevenueResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getRevenue(from, to)
      .then((res) => {
        if (!cancelled) setData(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load revenue data.')
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
          <h1 className="text-lg font-semibold">Revenue</h1>
          <p className="text-sm text-slate-500">Gross and net revenue over the selected period.</p>
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

      {isLoading && (
        <p className="text-sm text-slate-500">Loading revenue data...</p>
      )}

      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">
          {error}
        </div>
      )}

      {!isLoading && !error && data && (
        <>
          <div className="grid grid-cols-4 gap-4 mb-8">
            <SummaryCard label="Gross Revenue" value={formatCurrency(data.summary.total_gross_revenue)} />
            <SummaryCard label="Net Revenue" value={formatCurrency(data.summary.total_net_revenue)} />
            <SummaryCard label="Orders" value={data.summary.total_orders.toLocaleString('en-IN')} />
            <SummaryCard label="Avg Order Value" value={formatCurrency(data.summary.average_order_value)} />
          </div>

          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Date</th>
                  <th className="px-4 py-2 font-medium text-right">Gross Revenue</th>
                  <th className="px-4 py-2 font-medium text-right">Net Revenue</th>
                  <th className="px-4 py-2 font-medium text-right">Orders</th>
                </tr>
              </thead>
              <tbody>
                {data.daily.length === 0 && (
                  <tr>
                    <td colSpan={4} className="px-4 py-6 text-center text-slate-500">
                      No revenue data for this period.
                    </td>
                  </tr>
                )}
                {data.daily.map((row) => (
                  <tr key={row.date} className="border-t border-slate-800">
                    <td className="px-4 py-2">{row.date}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.gross_revenue)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.net_revenue)}</td>
                    <td className="px-4 py-2 text-right">{row.orders_count}</td>
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

function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-slate-800 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className="text-lg font-semibold">{value}</p>
    </div>
  )
}
