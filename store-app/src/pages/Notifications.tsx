import { useEffect, useState, useCallback } from 'react'
import { listWarehouseNotifications, markNotificationRead, markAllNotificationsRead } from '../api/warehouse'
import type { WarehouseNotification } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

const TYPE_LABELS: Record<string, string> = {
  new_order: 'New Order',
  urgent_order: 'Urgent Order',
  order_cancelled: 'Order Cancelled',
  low_stock: 'Low Stock',
  out_of_stock: 'Out of Stock',
  expiry_alert: 'Expiry Alert',
  stock_transfer: 'Stock Transfer',
  receiving: 'Receiving',
  handover_required: 'Handover Required',
  exception_created: 'Exception',
}

export default function Notifications() {
  const [notifications, setNotifications] = useState<WarehouseNotification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [filter, setFilter] = useState<'all' | 'unread'>('all')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    setIsLoading(true)
    try {
      const data = await listWarehouseNotifications({
        is_read: filter === 'unread' ? false : undefined,
        page,
        limit: 25,
      })
      setNotifications(data.notifications)
      setUnreadCount(data.unread_count)
      setTotalPages(data.total_pages || 1)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load notifications.'))
    } finally {
      setIsLoading(false)
    }
  }, [filter, page])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    setPage(1)
  }, [filter])

  async function handleMarkRead(id: number) {
    try {
      await markNotificationRead(id)
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to mark as read.'))
    }
  }

  async function handleMarkAllRead() {
    try {
      await markAllNotificationsRead()
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to mark all as read.'))
    }
  }

  return (
    <div className="p-6 max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-display text-2xl font-semibold">
          Notifications
          {unreadCount > 0 && (
            <span className="ml-2 text-xs bg-amber-500 text-white rounded-full px-2 py-0.5 align-middle">
              {unreadCount} unread
            </span>
          )}
        </h1>
        <div className="flex gap-2">
          <button
            onClick={handleMarkAllRead}
            disabled={unreadCount === 0}
            className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 transition-colors"
          >
            Mark all read
          </button>
          <button
            onClick={load}
            className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
          >
            Refresh
          </button>
        </div>
      </div>

      <div className="flex gap-1 border-b border-slate-800 mb-4">
        {(['all', 'unread'] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-2 text-sm capitalize border-b-2 transition-colors ${
              filter === f
                ? 'border-amber-400 text-amber-300 font-medium'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading notifications...</p>}

      {!isLoading && notifications.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No notifications.
        </div>
      )}

      {!isLoading && notifications.length > 0 && (
        <div className="space-y-2">
          {notifications.map((n) => (
            <div
              key={n.id}
              className={`border rounded-xl p-4 flex items-start justify-between gap-4 ${
                n.is_read ? 'border-slate-800 bg-slate-900/50' : 'border-amber-800 bg-amber-950/20'
              }`}
            >
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400">
                    {TYPE_LABELS[n.type] ?? n.type}
                  </span>
                  {!n.is_read && <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />}
                </div>
                <p className="text-sm font-medium">{n.title}</p>
                {n.message && <p className="text-xs text-slate-400 mt-0.5">{n.message}</p>}
                <p className="text-xs text-slate-600 mt-1">{new Date(n.created_at).toLocaleString()}</p>
              </div>
              {!n.is_read && (
                <button
                  onClick={() => handleMarkRead(n.id)}
                  className="text-xs px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors whitespace-nowrap"
                >
                  Mark read
                </button>
              )}
            </div>
          ))}
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
