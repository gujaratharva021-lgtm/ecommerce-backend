import { useCallback, useEffect, useState } from 'react'
import { listStockMovements } from '../api/warehouse'
import type { StockMovement } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

const MOVEMENT_TYPES = [
  { value: '', label: 'All' },
  { value: 'sale', label: 'Sale' },
  { value: 'adjustment', label: 'Adjustment' },
  { value: 'receiving', label: 'Receiving' },
  { value: 'transfer', label: 'Transfer' },
]

export default function StockMovements() {
  const [movements, setMovements] = useState<StockMovement[]>([])
  const [movementType, setMovementType] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listStockMovements({ movement_type: movementType || undefined, page, limit: 30 })
      setMovements(res.movements ?? [])
      setTotalPages(res.total_pages ?? 1)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load stock movements.'))
    } finally {
      setIsLoading(false)
    }
  }, [movementType, page])

  useEffect(() => { load() }, [load])

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-1">
        <h1 className="font-display text-2xl font-semibold">Stock Movements</h1>
      </div>
      <p className="text-sm text-slate-400 mb-5">History of every inventory change in your warehouse.</p>

      <div className="flex gap-2 mb-4">
        {MOVEMENT_TYPES.map((t) => (
          <button
            key={t.value}
            onClick={() => { setMovementType(t.value); setPage(1) }}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium ${
              movementType === t.value ? 'bg-red-500 text-white' : 'bg-slate-800 text-slate-400 hover:text-slate-200'
            }`}
          >
            {t.label}
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
      ) : movements.length === 0 ? (
        <p className="text-sm text-slate-500">No stock movements found.</p>
      ) : (
        <>
          <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-slate-500 border-b border-slate-800">
                  <th className="px-4 py-2.5 font-medium">Product</th>
                  <th className="px-4 py-2.5 font-medium">Type</th>
                  <th className="px-4 py-2.5 font-medium">Previous</th>
                  <th className="px-4 py-2.5 font-medium">Change</th>
                  <th className="px-4 py-2.5 font-medium">New Qty</th>
                  <th className="px-4 py-2.5 font-medium">Reason</th>
                  <th className="px-4 py-2.5 font-medium">Date</th>
                </tr>
              </thead>
              <tbody>
                {movements.map((m) => (
                  <tr key={m.id} className="border-b border-slate-800/60 last:border-0">
                    <td className="px-4 py-2.5 text-slate-200">{m.product?.name ?? `#${m.product_id}`}</td>
                    <td className="px-4 py-2.5">
                      <span className="px-2 py-0.5 rounded-full text-xs bg-slate-800 text-slate-300 capitalize">
                        {m.movement_type}
                      </span>
                    </td>
                    <td className="px-4 py-2.5 text-slate-400">{m.previous_qty}</td>
                    <td className={`px-4 py-2.5 font-medium ${m.change >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                      {m.change >= 0 ? '+' : ''}{m.change}
                    </td>
                    <td className="px-4 py-2.5 text-slate-200 font-medium">{m.new_qty}</td>
                    <td className="px-4 py-2.5 text-slate-400 text-xs">{m.reason ?? '-'}</td>
                    <td className="px-4 py-2.5 text-slate-500 text-xs">
                      {(m as any).created_at ? new Date((m as any).created_at).toLocaleString() : '-'}
                    </td>
                  </tr>
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
