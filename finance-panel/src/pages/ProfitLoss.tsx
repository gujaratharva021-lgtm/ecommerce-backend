import { useEffect, useState } from 'react'
import { getProfitLoss } from '../api/finance'
import type { ProfitLoss } from '../types/finance'

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

export default function ProfitLossPage() {
  const [from, setFrom] = useState(daysAgoISO(30))
  const [to, setTo] = useState(todayISO())
  const [data, setData] = useState<ProfitLoss | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getProfitLoss(from, to)
      .then((res) => {
        if (!cancelled) setData(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load profit & loss data.')
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
          <h1 className="text-lg font-semibold">Profit &amp; Loss</h1>
          <p className="text-sm text-slate-500">Revenue, cost of goods, and operating expenses for the selected period.</p>
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

      {isLoading && <p className="text-sm text-slate-500">Loading profit &amp; loss data...</p>}

      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && data && (
        <div className="max-w-xl">
          {data.cost_price_coverage < 100 && (
            <div className="border border-amber-900 bg-amber-950/30 text-amber-300 text-xs rounded-lg px-4 py-3 mb-6">
              {data.cost_price_coverage === 0
                ? 'No product cost prices are set yet, so COGS and Gross Profit below are not meaningful — set cost price on products to enable an accurate P&L.'
                : `Only ${data.cost_price_coverage.toFixed(0)}% of sold items had a cost price set for this period, so COGS is understated.`}
            </div>
          )}

          <div className="border border-slate-800 rounded-xl divide-y divide-slate-800">
            <PLRow label="Revenue" value={data.revenue} />
            <PLRow label="Cost of Goods Sold (COGS)" value={-data.cogs} isSubtractor />
            <PLRow label="Gross Profit" value={data.gross_profit} isBold />
            <PLRow label="Operating Expenses" value={-data.operating_expenses} isSubtractor />
            <PLRow label="EBITDA" value={data.ebitda} isBold />
            <PLRow label="Net Profit" value={data.net_profit} isBold isFinal />
          </div>
        </div>
      )}
    </div>
  )
}

function PLRow({
  label,
  value,
  isBold = false,
  isSubtractor = false,
  isFinal = false,
}: {
  label: string
  value: number
  isBold?: boolean
  isSubtractor?: boolean
  isFinal?: boolean
}) {
  return (
    <div
      className={`flex items-center justify-between px-4 py-3 ${
        isFinal ? 'bg-slate-900' : ''
      }`}
    >
      <span className={isBold ? 'font-semibold' : 'text-slate-400'}>{label}</span>
      <span
        className={`${isBold ? 'font-semibold' : ''} ${
          value < 0 ? 'text-red-400' : isFinal ? (value >= 0 ? 'text-emerald-400' : 'text-red-400') : ''
        }`}
      >
        {isSubtractor && value !== 0 ? '(' + formatCurrency(Math.abs(value)) + ')' : formatCurrency(value)}
      </span>
    </div>
  )
}
