import { useEffect, useState, useCallback } from 'react'
import { getStaffOverview } from '../api/warehouse'
import type { StaffOverviewRow } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

const ROLE_LABELS: Record<string, string> = {
  picker: 'Picker',
  packer: 'Packer',
  inventory_staff: 'Inventory Staff',
  supervisor: 'Supervisor',
  warehouse_manager: 'Warehouse Manager',
}

function timeAgo(dateStr?: string | null): string {
  if (!dateStr) return 'No activity yet'
  const ms = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(ms / 60000)
  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

function initials(name: string): string {
  const parts = name.trim().split(/\s+/)
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

const IconRefresh = (
  <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"/><path d="M8 16H3v5"/></svg>
)

export default function Staff() {
  const [rows, setRows] = useState<StaffOverviewRow[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    try {
      const data = await getStaffOverview()
      setRows(data.staff)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load staff.'))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [load])

  const activeCount = rows.filter((s) => s.is_active).length
  const totalOrders = rows.reduce((sum, s) => sum + s.orders_handled, 0)

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-1">
        <div>
          <p className="font-mono text-[10px] tracking-widest text-amber-500 uppercase mb-1">Team</p>
          <h1 className="font-display text-2xl font-semibold">Staff</h1>
        </div>
        <button
          onClick={load}
          className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 border border-slate-700 transition-colors flex items-center gap-1.5"
        >
          {IconRefresh}
          Refresh
        </button>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mt-6">
          {error}
        </div>
      )}

      {isLoading && <p className="text-sm text-slate-400 mt-6">Loading staff...</p>}

      {!isLoading && rows.length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mt-6 mb-6">
          <div className="relative border border-slate-800 bg-gradient-to-b from-slate-900 to-slate-900/60 rounded-xl p-4 ring-1 ring-slate-700/30">
            <p className="text-xs text-slate-400 mb-2">Total Staff</p>
            <p className="font-mono text-2xl font-semibold text-slate-100">{rows.length}</p>
          </div>
          <div className="relative border border-slate-800 bg-gradient-to-b from-slate-900 to-slate-900/60 rounded-xl p-4 ring-1 ring-emerald-500/20">
            <p className="text-xs text-slate-400 mb-2">Active Now</p>
            <p className="font-mono text-2xl font-semibold text-emerald-300">{activeCount}</p>
          </div>
          <div className="relative border border-slate-800 bg-gradient-to-b from-slate-900 to-slate-900/60 rounded-xl p-4 ring-1 ring-amber-500/20">
            <p className="text-xs text-slate-400 mb-2">Orders Handled</p>
            <p className="font-mono text-2xl font-semibold text-amber-300">{totalOrders}</p>
          </div>
        </div>
      )}

      {!isLoading && rows.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500 mt-6">
          No staff members found for your warehouse.
        </div>
      )}

      {!isLoading && rows.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Staff</th>
                <th className="text-left px-4 py-2.5">Role</th>
                <th className="text-left px-4 py-2.5">Status</th>
                <th className="text-left px-4 py-2.5">Current Task</th>
                <th className="text-right px-4 py-2.5">Orders Handled</th>
                <th className="text-left px-4 py-2.5">Last Activity</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {rows.map((s) => (
                <tr key={s.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-amber-400 to-orange-600 flex items-center justify-center text-[11px] font-semibold text-white shrink-0">
                        {initials(s.name)}
                      </div>
                      <div>
                        <p className="font-medium">{s.name}</p>
                        <p className="text-xs text-slate-500">{s.phone}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-xs px-2 py-1 rounded-full bg-amber-500/15 text-amber-300">
                      {ROLE_LABELS[s.role] ?? s.role}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`text-xs px-2 py-1 rounded-full inline-flex items-center gap-1.5 ${
                        s.is_active ? 'bg-emerald-500/15 text-emerald-400' : 'bg-slate-700 text-slate-300'
                      }`}
                    >
                      <span className={`w-1.5 h-1.5 rounded-full ${s.is_active ? 'bg-emerald-400' : 'bg-slate-500'}`} />
                      {s.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-300">{s.current_task ?? '-'}</td>
                  <td className="px-4 py-3 text-right font-mono text-slate-200">{s.orders_handled}</td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{timeAgo(s.last_activity)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
