import { useEffect, useState, useCallback } from 'react'
import { listAuditLogs } from '../api/warehouse'
import type { WarehouseAuditLog } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

const ACTION_LABELS: Record<string, string> = {
  stock_adjustment: 'Stock Adjustment',
  accept_order: 'Accept Order',
  mark_pick_item: 'Mark Pick Item',
  complete_picking: 'Complete Picking',
  complete_packing: 'Complete Packing',
  handover_order: 'Handover Order',
  delete_zone: 'Delete Zone',
  delete_rack: 'Delete Rack',
  delete_bin: 'Delete Bin',
}

export default function AuditLogs() {
  const [logs, setLogs] = useState<WarehouseAuditLog[]>([])
  const [actionFilter, setActionFilter] = useState('')
  const [entityTypeFilter, setEntityTypeFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    setIsLoading(true)
    try {
      const data = await listAuditLogs({
        action: actionFilter || undefined,
        entity_type: entityTypeFilter || undefined,
        page,
        limit: 25,
      })
      setLogs(data.audit_logs)
      setTotalPages(data.total_pages || 1)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load audit logs.'))
    } finally {
      setIsLoading(false)
    }
  }, [actionFilter, entityTypeFilter, page])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    setPage(1)
  }, [actionFilter, entityTypeFilter])

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-display text-2xl font-semibold">Audit Logs</h1>
        <button
          onClick={load}
          className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          Refresh
        </button>
      </div>

      <div className="flex flex-wrap gap-2 mb-4">
        <select
          value={actionFilter}
          onChange={(e) => setActionFilter(e.target.value)}
          className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-slate-300"
        >
          <option value="">All actions</option>
          {Object.entries(ACTION_LABELS).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
        <select
          value={entityTypeFilter}
          onChange={(e) => setEntityTypeFilter(e.target.value)}
          className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-slate-300"
        >
          <option value="">All entity types</option>
          <option value="order">Order</option>
          <option value="inventory">Inventory</option>
          <option value="picking_task_item">Picking Item</option>
          <option value="warehouse_zone">Zone</option>
          <option value="warehouse_rack">Rack</option>
          <option value="warehouse_bin">Bin</option>
        </select>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading audit logs...</p>}

      {!isLoading && logs.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No audit log entries found for this filter.
        </div>
      )}

      {!isLoading && logs.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Action</th>
                <th className="text-left px-4 py-2.5">Entity</th>
                <th className="text-left px-4 py-2.5">Staff</th>
                <th className="text-left px-4 py-2.5">Before</th>
                <th className="text-left px-4 py-2.5">After</th>
                <th className="text-left px-4 py-2.5">Timestamp</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {logs.map((log) => (
                <tr key={log.id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3 font-medium text-indigo-300">
                    {ACTION_LABELS[log.action] ?? log.action}
                  </td>
                  <td className="px-4 py-3 text-slate-400">
                    {log.entity_type}
                    <span className="text-slate-600 ml-1.5 text-xs">#{log.entity_id}</span>
                  </td>
                  <td className="px-4 py-3">{log.staff_name}</td>
                  <td className="px-4 py-3 text-slate-500 text-xs max-w-[180px] truncate" title={log.before_value}>
                    {log.before_value || '-'}
                  </td>
                  <td className="px-4 py-3 text-slate-400 text-xs max-w-[220px] truncate" title={log.after_value}>
                    {log.after_value || '-'}
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">
                    {new Date(log.created_at).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!isLoading && totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="text-xs text-slate-500">
            Page {page} of {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
