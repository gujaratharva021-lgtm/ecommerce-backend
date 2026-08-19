import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { listReturns, approveReturn, rejectReturn } from '../api/admin'
import type { ReturnRequest } from '../types/admin'

export default function Returns() {
  const [returns, setReturns] = useState<ReturnRequest[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actingId, setActingId] = useState<number | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listReturns()
      setReturns(res.return_requests ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load returns.')
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
      await approveReturn(id)
      setReturns((prev) =>
        prev.map((r) => (r.id === id ? { ...r, status: 'approved' } : r))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to approve return.')
    } finally {
      setActingId(null)
    }
  }

  async function handleReject(id: number) {
    setActingId(id)
    try {
      await rejectReturn(id)
      setReturns((prev) =>
        prev.map((r) => (r.id === id ? { ...r, status: 'rejected' } : r))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to reject return.')
    } finally {
      setActingId(null)
    }
  }

  function statusColor(status: string) {
    if (status === 'approved') return 'bg-emerald-500/15 text-emerald-400'
    if (status === 'rejected') return 'bg-red-500/15 text-red-400'
    return 'bg-amber-500/15 text-amber-400'
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Returns</h1>
          <p className="text-sm text-slate-400 mt-1">
            {returns.length} return{returns.length !== 1 ? 's' : ''}
          </p>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && returns.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No return requests yet.
          </div>
        )}

        {!isLoading && returns.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Order ID</th>
                  <th className="px-4 py-3 font-medium">Product</th>
                  <th className="px-4 py-3 font-medium">Reason</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {returns.map((r) => (
                  <tr key={r.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">#{r.order_id}</td>
                    <td className="px-4 py-3 text-slate-400">
                      {r.items && r.items.length > 0
                        ? r.items.map((it) => it.order_item?.product?.name ?? 'Unknown').join(', ')
                        : '-'}
                    </td>
                    <td className="px-4 py-3 text-slate-400">{r.reason ?? '-'}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-1 rounded-full ${statusColor(r.status)}`}>
                        {r.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right space-x-3">
                      {r.status === 'pending' && (
                        <>
                          <button
                            onClick={() => handleApprove(r.id)}
                            disabled={actingId === r.id}
                            className="text-emerald-400 hover:text-emerald-300 text-xs disabled:opacity-50"
                          >
                            Approve
                          </button>
                          <button
                            onClick={() => handleReject(r.id)}
                            disabled={actingId === r.id}
                            className="text-red-400 hover:text-red-300 text-xs disabled:opacity-50"
                          >
                            Reject
                          </button>
                        </>
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
