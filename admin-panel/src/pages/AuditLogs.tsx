import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { getAuditLogs } from '../api/admin'
import type { AuditLog } from '../types/admin'

export default function AuditLogs() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)

  const [actionFilter, setActionFilter] = useState('')
  const [entityTypeFilter, setEntityTypeFilter] = useState('')

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const params: Record<string, any> = { page, limit: 50 }
      if (actionFilter) params.action = actionFilter
      if (entityTypeFilter) params.entity_type = entityTypeFilter
      const res = await getAuditLogs(params)
      setLogs(res.logs ?? [])
      setTotalPages(res.total_pages ?? 1)
      setTotal(res.total ?? 0)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load audit logs.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, actionFilter, entityTypeFilter])

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Audit Logs</h1>
          <p className="text-sm text-slate-400 mt-1">
            {total} recorded action{total !== 1 ? 's' : ''} &middot; every sensitive admin action is logged here
          </p>
        </div>

        <div className="flex items-center gap-3 mb-6">
          <input
            type="text"
            placeholder="Filter by action (e.g. update_settings)"
            value={actionFilter}
            onChange={(e) => {
              setActionFilter(e.target.value)
              setPage(1)
            }}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm w-72"
          />
          <input
            type="text"
            placeholder="Filter by entity type (e.g. product)"
            value={entityTypeFilter}
            onChange={(e) => {
              setEntityTypeFilter(e.target.value)
              setPage(1)
            }}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm w-72"
          />
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && logs.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No audit log entries found.
          </div>
        )}

        {!isLoading && logs.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Time</th>
                  <th className="px-4 py-3 font-medium">Admin</th>
                  <th className="px-4 py-3 font-medium">Action</th>
                  <th className="px-4 py-3 font-medium">Entity</th>
                  <th className="px-4 py-3 font-medium">Details</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <tr key={log.id} className="border-t border-slate-800">
                    <td className="px-4 py-3 text-slate-400 whitespace-nowrap">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-slate-300">{log.admin_phone}</td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-1 rounded-md text-xs font-medium bg-indigo-500/15 text-indigo-300">
                        {log.action}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-400">
                      {log.entity_type}
                      {log.entity_id && log.entity_id !== '-' ? ` #${log.entity_id}` : ''}
                    </td>
                    <td className="px-4 py-3 text-slate-400">{log.details}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!isLoading && totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-6">
            <button
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-sm disabled:opacity-40"
            >
              Prev
            </button>
            <span className="text-sm text-slate-400">
              Page {page} of {totalPages}
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-sm disabled:opacity-40"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </Layout>
  )
}
