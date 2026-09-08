import { useCallback, useEffect, useState } from 'react'
import { listAuditLogs } from '../api/warehouse'
import type { WarehouseAuditLog } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

export default function AuditLogs() {
  const [logs, setLogs] = useState<WarehouseAuditLog[]>([])
  const [entityType, setEntityType] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listAuditLogs({ entity_type: entityType || undefined, page, limit: 30 })
      setLogs(res.audit_logs ?? [])
      setTotalPages(res.total_pages ?? 1)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load audit logs.'))
    } finally {
      setIsLoading(false)
    }
  }, [entityType, page])

  useEffect(() => { load() }, [load])

  const entityTypes = ['', 'warehouse_zone', 'warehouse_rack', 'warehouse_bin', 'batch', 'inventory', 'order']

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-1">
        <h1 className="font-display text-2xl font-semibold">Audit Logs</h1>
      </div>
      <p className="text-sm text-slate-400 mb-5">Every change made in your warehouse, by whom and when.</p>

      <div className="flex gap-2 mb-4 flex-wrap">
        {entityTypes.map((t) => (
          <button
            key={t}
            onClick={() => { setEntityType(t); setPage(1) }}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium capitalize ${
              entityType === t ? 'bg-red-500 text-white' : 'bg-slate-800 text-slate-400 hover:text-slate-200'
            }`}
          >
            {t === '' ? 'All' : t.replace('_', ' ')}
          </button>
        ))}
      </div>

      {error && (
        <div className="mb-4 px-4 py-2 rounded-lg bg-red-500/10 border border-red-500/30 text-red-300 text-sm">
          {error}
        </div>
      )}

      {isLoading ? (
        <p className="text-sm text-slate-400">Loading...</p>
      ) : logs.length === 0 ? (
        <p className="text-sm text-slate-500">No audit logs found.</p>
      ) : (
        <>
          <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-slate-500 border-b border-slate-800">
                  <th className="px-4 py-2.5 font-medium">Staff</th>
                  <th className="px-4 py-2.5 font-medium">Action</th>
                  <th className="px-4 py-2.5 font-medium">Entity</th>
                  <th className="px-4 py-2.5 font-medium">Date</th>
                  <th className="px-4 py-2.5 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <>
                    <tr key={log.id} className="border-b border-slate-800/60 last:border-0">
                      <td className="px-4 py-2.5 text-slate-200">{log.staff_name}</td>
                      <td className="px-4 py-2.5">
                        <span className="px-2 py-0.5 rounded-full text-xs bg-slate-800 text-slate-300 capitalize">
                          {log.action.replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td className="px-4 py-2.5 text-slate-400 text-xs">
                        {log.entity_type.replace(/_/g, ' ')} #{log.entity_id}
                      </td>
                      <td className="px-4 py-2.5 text-slate-500 text-xs">{new Date(log.created_at).toLocaleString()}</td>
                      <td className="px-4 py-2.5 text-right">
                        {(log.before_value || log.after_value) && (
                          <button
                            onClick={() => setExpandedId(expandedId === log.id ? null : log.id)}
                            className="text-xs text-red-400 hover:text-red-300"
                          >
                            {expandedId === log.id ? 'Hide' : 'View'}
                          </button>
                        )}
                      </td>
                    </tr>
                    {expandedId === log.id && (
                      <tr className="border-b border-slate-800/60 last:border-0 bg-slate-950/60">
                        <td colSpan={5} className="px-4 py-3">
                          <div className="grid grid-cols-2 gap-4 text-xs">
                            <div>
                              <p className="text-slate-500 mb-1">Before</p>
                              <pre className="bg-slate-900 border border-slate-800 rounded-lg p-2 text-slate-400 whitespace-pre-wrap break-all">
                                {log.before_value || '-'}
                              </pre>
                            </div>
                            <div>
                              <p className="text-slate-500 mb-1">After</p>
                              <pre className="bg-slate-900 border border-slate-800 rounded-lg p-2 text-slate-400 whitespace-pre-wrap break-all">
                                {log.after_value || '-'}
                              </pre>
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </>
                ))}
              </tbody>
            </table>
          </div>
          {totalPages > 1 && (
            <div className="flex items-center justify-end gap-2 mt-4">
              <button
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
                className="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 disabled:opacity-40"
              >
                Previous
              </button>
              <span className="text-xs text-slate-500">Page {page} of {totalPages}</span>
              <button
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 disabled:opacity-40"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
