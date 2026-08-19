import { useEffect, useState } from 'react'
import { getRevenue } from '../api/finance'
import type { RevenueSummary } from '../types/finance'

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
  const [data, setData] = useState<RevenueSummary | null>(null)
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
          <div className="grid grid-cols-3 gap-4 mb-4">
            <SummaryCard label="Gross Sales" value={formatCurrency(data.gross_sales)} />
            <SummaryCard label="Net Sales" value={formatCurrency(data.net_sales)} />
            <SummaryCard label="Discounts" value={formatCurrency(data.discounts)} />
          </div>
          <div className="grid grid-cols-3 gap-4 mb-8">
            <SummaryCard label="Delivery Charge" value={formatCurrency(data.delivery_charge)} />
            <SummaryCard label="Platform Fee" value={formatCurrency(data.platform_fee)} />
            <SummaryCard label="Orders" value={data.order_count.toLocaleString('en-IN')} />
          </div>

          <h2 className="text-sm font-semibold mb-3">Revenue by Warehouse</h2>
          <div className="border border-slate-800 rounded-xl overflow-hidden mb-8">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Warehouse</th>
                  <th className="px-4 py-2 font-medium text-right">Revenue</th>
                  <th className="px-4 py-2 font-medium text-right">Orders</th>
                </tr>
              </thead>
              <tbody>
                {data.by_warehouse.length === 0 && (
                  <tr>
                    <td colSpan={3} className="px-4 py-6 text-center text-slate-500">
                      No warehouse data for this period.
                    </td>
                  </tr>
                )}
                {data.by_warehouse.map((row) => (
                  <tr key={row.warehouse_id ?? 'unassigned'} className="border-t border-slate-800">
                    <td className="px-4 py-2">{row.warehouse_name}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.revenue)}</td>
                    <td className="px-4 py-2 text-right">{row.order_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <h2 className="text-sm font-semibold mb-3">Top Products by Revenue</h2>
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Product</th>
                  <th className="px-4 py-2 font-medium text-right">Revenue</th>
                  <th className="px-4 py-2 font-medium text-right">Quantity</th>
                </tr>
              </thead>
              <tbody>
                {data.by_product.length === 0 && (
                  <tr>
                    <td colSpan={3} className="px-4 py-6 text-center text-slate-500">
                      No product data for this period.
                    </td>
                  </tr>
                )}
                {data.by_product.map((row) => (
                  <tr key={row.product_id} className="border-t border-slate-800">
                    <td className="px-4 py-2">{row.product_name}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.revenue)}</td>
                    <td className="px-4 py-2 text-right">{row.quantity}</td>
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
