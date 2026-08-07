import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import {
  listStockTransfers,
  approveStockTransfer,
  rejectStockTransfer,
  cancelStockTransfer,
} from '../api/admin'
import type { StockTransfer } from '../types/admin'

export default function StockTransfers() {
  const [transfers, setTransfers] = useState<StockTransfer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actingId, setActingId] = useState<number | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listStockTransfers()
      setTransfers(res.stock_transfers ?? res.transfers ?? res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load stock transfers.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function handleApprove(id: number) {
    setActingId(id)
    try {
      await approveStockTransfer(id)
      setTransfers((prev) =>
        prev.map((t) => (t.id === id ? { ...t, status: 'in_transit' } : t))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to approve transfer.')
    } finally {
      setActingId(null)
    }
  }

  async function handleReject(id: number) {
    setActingId(id)
    try {
      await rejectStockTransfer(id)
      setTransfers((prev) =>
        prev.map((t) => (t.id === id ? { ...t, status: 'rejected' } : t))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to reject transfer.')
    } finally {
      setActingId(null)
    }
  }

  async function handleCancel(id: number) {
    if (!confirm('Cancel this transfer? Stock will be restored to the source warehouse.')) return
    setActingId(id)
    try {
      await cancelStockTransfer(id)
      setTransfers((prev) =>
        prev.map((t) => (t.id === id ? { ...t, status: 'cancelled' } : t))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to cancel transfer.')
    } finally {
      setActingId(null)
    }
  }

  function statusColor(status: string) {
    if (status === 'in_transit') return 'bg-blue-500/15 text-blue-400'
    if (status === 'received') return 'bg-emerald-500/15 text-emerald-400'
    if (status === 'cancelled') return 'bg-slate-500/15 text-slate-400'
    if (status === 'rejected') return 'bg-red-500/15 text-red-400'
    return 'bg-amber-500/15 text-amber-400'
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Stock Transfers</h1>
          <p className="text-sm text-slate-400 mt-1">
            {transfers.length} transfer{transfers.length !== 1 ? 's' : ''}
          </p>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && transfers.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No stock transfer requests yet.
          </div>
        )}

        {!isLoading && transfers.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Product ID</th>
                  <th className="px-4 py-3 font-medium">From</th>
                  <th className="px-4 py-3 font-medium">To</th>
                  <th className="px-4 py-3 font-medium">Quantity</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {transfers.map((t) => (
                  <tr key={t.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">#{t.product_id}</td>
                    <td className="px-4 py-3">#{t.from_warehouse_id}</td>
                    <td className="px-4 py-3">#{t.to_warehouse_id}</td>
                    <td className="px-4 py-3">{t.quantity}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-1 rounded-full ${statusColor(t.status)}`}>
                        {t.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right space-x-3">
                      {t.status === 'pending' && (
                        <>
                          <button
                            onClick={() => handleApprove(t.id)}
                            disabled={actingId === t.id}
                            className="text-emerald-400 hover:text-emerald-300 text-xs disabled:opacity-50"
                          >
                            Approve
                          </button>
                          <button
                            onClick={() => handleReject(t.id)}
                            disabled={actingId === t.id}
                            className="text-red-400 hover:text-red-300 text-xs disabled:opacity-50"
                          >
                            Reject
                          </button>
                        </>
                      )}
                      {t.status === 'in_transit' && (
                        <button
                          onClick={() => handleCancel(t.id)}
                          disabled={actingId === t.id}
                          className="text-red-400 hover:text-red-300 text-xs disabled:opacity-50"
                        >
                          Cancel
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </Layout>
  )
}
