import { useEffect, useState, useCallback } from 'react'
import { listExceptions, updateException } from '../api/warehouse'
import type { WarehouseException, ExceptionStatus } from '../types/warehouse'
import StatusBadge from '../components/StatusBadge'
import { getErrorMessage } from '../utils/errors'

const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'open', label: 'Open' },
  { value: 'investigating', label: 'Investigating' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'closed', label: 'Closed' },
]

export default function Exceptions() {
  const [exceptions, setExceptions] = useState<WarehouseException[]>([])
  const [statusFilter, setStatusFilter] = useState('open')
  const [priorityFilter, setPriorityFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selected, setSelected] = useState<WarehouseException | null>(null)
  const [resolution, setResolution] = useState('')
  const [actionStatus, setActionStatus] = useState<ExceptionStatus>('resolved')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const load = useCallback(async () => {
    setError(null)
    setIsLoading(true)
    try {
      const data = await listExceptions({
        status: statusFilter || undefined,
        priority: priorityFilter || undefined,
        page,
        limit: 20,
      })
      setExceptions(data.exceptions)
      setTotalPages(data.total_pages || 1)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load exceptions.'))
    } finally {
      setIsLoading(false)
    }
  }, [statusFilter, priorityFilter, page])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    setPage(1)
  }, [statusFilter, priorityFilter])

  const openResolveDialog = (exc: WarehouseException, status: ExceptionStatus) => {
    setSelected(exc)
    setActionStatus(status)
    setResolution('')
  }

  const submitUpdate = async () => {
    if (!selected) return
    if ((actionStatus === 'resolved' || actionStatus === 'closed') && !resolution.trim()) {
      setError('Resolution note is required to resolve or close an exception.')
      return
    }
    setIsSubmitting(true)
    setError(null)
    try {
      await updateException(selected.id, { status: actionStatus, resolution: resolution.trim() || undefined })
      setSelected(null)
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to update exception.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-display text-2xl font-semibold">Exceptions</h1>
        <button
          onClick={load}
          className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          Refresh
        </button>
      </div>

      <div className="flex flex-wrap gap-2 mb-4">
        <div className="flex gap-1">
          {STATUS_FILTERS.map((f) => (
            <button
              key={f.value}
              onClick={() => setStatusFilter(f.value)}
              className={`text-xs px-3 py-1.5 rounded-lg transition-colors ${
                statusFilter === f.value
                  ? 'bg-amber-500/20 text-amber-300'
                  : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
        <select
          value={priorityFilter}
          onChange={(e) => setPriorityFilter(e.target.value)}
          className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-slate-300"
        >
          <option value="">All priorities</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading exceptions...</p>}

      {!isLoading && exceptions.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No exceptions found for this filter.
        </div>
      )}

      {!isLoading && exceptions.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Order</th>
                <th className="text-left px-4 py-2.5">Product</th>
                <th className="text-left px-4 py-2.5">Type</th>
                <th className="text-left px-4 py-2.5">Reason</th>
                <th className="text-left px-4 py-2.5">Priority</th>
                <th className="text-left px-4 py-2.5">Status</th>
                <th className="text-left px-4 py-2.5">Created</th>
                <th className="text-right px-4 py-2.5">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {exceptions.map((exc) => (
                <tr key={exc.id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3">#{exc.order_id}</td>
                  <td className="px-4 py-3">{exc.product?.name ?? '-'}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={exc.type} />
                  </td>
                  <td className="px-4 py-3 text-slate-400 max-w-[220px] truncate" title={exc.reason}>
                    {exc.reason || '-'}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={exc.priority} />
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={exc.status} />
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">
                    {new Date(exc.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {(exc.status === 'open' || exc.status === 'investigating') && (
                      <div className="flex gap-1.5 justify-end">
                        {exc.status === 'open' && (
                          <button
                            onClick={() => openResolveDialog(exc, 'investigating')}
                            className="text-xs px-2.5 py-1 rounded-md bg-amber-500/15 text-amber-300 hover:bg-amber-500/25"
                          >
                            Investigate
                          </button>
                        )}
                        <button
                          onClick={() => openResolveDialog(exc, 'resolved')}
                          className="text-xs px-2.5 py-1 rounded-md bg-emerald-500/15 text-emerald-300 hover:bg-emerald-500/25"
                        >
                          Resolve
                        </button>
                        <button
                          onClick={() => openResolveDialog(exc, 'closed')}
                          className="text-xs px-2.5 py-1 rounded-md bg-slate-700 text-slate-300 hover:bg-slate-600"
                        >
                          Close
                        </button>
                      </div>
                    )}
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

      {selected && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 w-full max-w-md">
            <h2 className="text-sm font-semibold mb-1">
              Mark as {actionStatus.replace(/_/g, ' ')}
            </h2>
            <p className="text-xs text-slate-500 mb-4">
              Exception #{selected.id} &middot; Order #{selected.order_id}
            </p>
            {(actionStatus === 'resolved' || actionStatus === 'closed') && (
              <textarea
                value={resolution}
                onChange={(e) => setResolution(e.target.value)}
                placeholder="Resolution notes (required)"
                rows={3}
                className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 mb-4 focus:outline-none focus:border-amber-500"
              />
            )}
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setSelected(null)}
                disabled={isSubmitting}
                className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700"
              >
                Cancel
              </button>
              <button
                onClick={submitUpdate}
                disabled={isSubmitting}
                className="text-xs px-3 py-1.5 rounded-lg bg-amber-500/20 text-amber-300 hover:bg-amber-500/30 disabled:opacity-50"
              >
                {isSubmitting ? 'Saving...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
