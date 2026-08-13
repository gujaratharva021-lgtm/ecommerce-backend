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

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold">Staff</h1>
        <button
          onClick={load}
          className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          Refresh
        </button>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading staff...</p>}

      {!isLoading && rows.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
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
                <tr key={s.id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3">
                    <p className="font-medium">{s.name}</p>
                    <p className="text-xs text-slate-500">{s.phone}</p>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-xs px-2 py-1 rounded-full bg-indigo-500/15 text-indigo-300">
                      {ROLE_LABELS[s.role] ?? s.role}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`text-xs px-2 py-1 rounded-full ${
                        s.is_active ? 'bg-emerald-500/15 text-emerald-400' : 'bg-slate-700 text-slate-300'
                      }`}
                    >
                      {s.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-300">{s.current_task ?? '-'}</td>
                  <td className="px-4 py-3 text-right">{s.orders_handled}</td>
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
