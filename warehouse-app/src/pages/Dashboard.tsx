import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { getDashboard } from '../api/warehouse'
import type { WarehouseDashboardStats } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

function StatCard({
  label,
  value,
  suffix,
  tone = 'default',
}: {
  label: string
  value: number | string
  suffix?: string
  tone?: 'default' | 'warning' | 'danger' | 'success'
}) {
  const toneClass =
    tone === 'warning'
      ? 'text-amber-300'
      : tone === 'danger'
        ? 'text-rose-300'
        : tone === 'success'
          ? 'text-emerald-300'
          : 'text-slate-100'

  return (
    <div className="border border-slate-800 rounded-xl bg-slate-900 p-4">
      <p className="text-xs text-slate-400 mb-2">{label}</p>
      <p className={`text-2xl font-semibold ${toneClass}`}>
        {value}
        {suffix ? <span className="text-sm font-normal text-slate-500 ml-1">{suffix}</span> : null}
      </p>
    </div>
  )
}

export default function Dashboard() {
  const navigate = useNavigate()
  const [stats, setStats] = useState<WarehouseDashboardStats | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const load = useCallback(async () => {
    setError(null)
    try {
      const data = await getDashboard()
      setStats(data)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load dashboard.'))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    // Refresh every 30s so the dashboard stays near-real-time.
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [load])

  const alerts: { label: string; tone: 'warning' | 'danger'; onClick?: () => void }[] = []
  if (stats) {
    if (stats.out_of_stock_products > 0) {
      alerts.push({ label: `${stats.out_of_stock_products} product(s) out of stock`, tone: 'danger' })
    }
    if (stats.low_stock_products > 0) {
      alerts.push({ label: `${stats.low_stock_products} product(s) running low on stock`, tone: 'warning' })
    }
    if (stats.new_orders > 0) {
      alerts.push({
        label: `${stats.new_orders} new order(s) waiting to be accepted`,
        tone: 'warning',
        onClick: () => navigate('/orders?status=confirmed'),
      })
    }
    if (stats.pending_stock_transfers > 0) {
      alerts.push({
        label: `${stats.pending_stock_transfers} stock transfer(s) pending`,
        tone: 'warning',
        onClick: () => navigate('/stock-transfers'),
      })
    }
  }

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Dashboard</h1>
        <button
          onClick={load}
          className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          Refresh
        </button>
      </div>

      {isLoading && <p className="text-sm text-slate-400">Loading dashboard...</p>}
      {error && !isLoading && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {stats && (
        <>
          {alerts.length > 0 && (
            <div className="mb-6 space-y-2">
              {alerts.map((a, i) => (
                <button
                  key={i}
                  onClick={a.onClick}
                  disabled={!a.onClick}
                  className={`w-full text-left text-sm px-4 py-2.5 rounded-lg border flex items-center gap-2 ${
                    a.tone === 'danger'
                      ? 'border-rose-900 bg-rose-950/40 text-rose-300'
                      : 'border-amber-900 bg-amber-950/30 text-amber-300'
                  } ${a.onClick ? 'hover:brightness-125 cursor-pointer' : 'cursor-default'}`}
                >
                  {a.label}
                </button>
              ))}
            </div>
          )}

          <p className="text-xs uppercase tracking-wide text-slate-500 mb-2">Today</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
            <StatCard label="Today's Orders" value={stats.today_orders} />
            <StatCard label="New Orders" value={stats.new_orders} tone={stats.new_orders > 0 ? 'warning' : 'default'} />
            <StatCard label="Picking" value={stats.picking} />
            <StatCard label="Packed" value={stats.packed} />
            <StatCard label="Ready for Dispatch" value={stats.ready_for_dispatch} tone="success" />
            <StatCard label="Completed Today" value={stats.completed_today} tone="success" />
            <StatCard label="Cancelled Today" value={stats.cancelled_today} tone={stats.cancelled_today > 0 ? 'danger' : 'default'} />
            <StatCard label="Fulfillment Rate" value={stats.fulfillment_rate.toFixed(0)} suffix="%" tone="success" />
          </div>

          <p className="text-xs uppercase tracking-wide text-slate-500 mb-2">Performance</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
            <StatCard label="Avg Picking Time" value={stats.avg_picking_minutes.toFixed(1)} suffix="min" />
            <StatCard label="Avg Packing Time" value={stats.avg_packing_minutes.toFixed(1)} suffix="min" />
            <StatCard label="Active Staff" value={stats.active_staff} />
          </div>

          <p className="text-xs uppercase tracking-wide text-slate-500 mb-2">Inventory & Transfers</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <StatCard label="Low Stock Products" value={stats.low_stock_products} tone={stats.low_stock_products > 0 ? 'warning' : 'default'} />
            <StatCard label="Out of Stock Products" value={stats.out_of_stock_products} tone={stats.out_of_stock_products > 0 ? 'danger' : 'default'} />
            <StatCard label="Pending Stock Transfers" value={stats.pending_stock_transfers} tone={stats.pending_stock_transfers > 0 ? 'warning' : 'default'} />
          </div>
        </>
      )}
    </div>
  )
}
