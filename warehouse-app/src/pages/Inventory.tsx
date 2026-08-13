import { useEffect, useState, useCallback } from 'react'
import { getWarehouseInventory } from '../api/warehouse'
import type { WarehouseInventoryRow, WarehouseInventoryResponse } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

const TABS: { value: string; label: (data: WarehouseInventoryResponse | null) => string }[] = [
  { value: '', label: () => 'All' },
  { value: 'in_stock', label: (d) => `In Stock${d ? ` (${d.in_stock_count})` : ''}` },
  { value: 'low', label: (d) => `Low Stock${d ? ` (${d.low_stock_count})` : ''}` },
  { value: 'out', label: (d) => `Out of Stock${d ? ` (${d.out_of_stock_count})` : ''}` },
  { value: 'damaged', label: (d) => `Damaged${d ? ` (${d.damaged_count})` : ''}` },
  { value: 'expired', label: (d) => `Expired${d ? ` (${d.expired_count})` : ''}` },
]

function timeAgo(dateStr?: string | null): string {
  if (!dateStr) return ''
  const ms = Date.now() - new Date(dateStr).getTime()
  const days = Math.floor(ms / (1000 * 60 * 60 * 24))
  if (days < 1) return 'today'
  if (days === 1) return '1 day ago'
  return `${days} days ago`
}

export default function Inventory() {
  const [data, setData] = useState<WarehouseInventoryResponse | null>(null)
  const [rows, setRows] = useState<WarehouseInventoryRow[]>([])
  const [statusFilter, setStatusFilter] = useState('')
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [page, setPage] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    setIsLoading(true)
    try {
      const res = await getWarehouseInventory({
        search: search || undefined,
        stock_status: statusFilter || undefined,
        page,
        limit: 20,
      })
      setData(res)
      setRows(res.rows)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load inventory.'))
    } finally {
      setIsLoading(false)
    }
  }, [statusFilter, search, page])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    setPage(1)
  }, [statusFilter, search])

  const submitSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setSearch(searchInput.trim())
  }

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-2">
        <h1 className="text-lg font-semibold">Inventory</h1>
        <button onClick={load} className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors">
          Refresh
        </button>
      </div>
      {data && (
        <p className="text-xs text-slate-500 mb-4">
          Low-stock threshold: {data.low_stock_threshold} units &middot; {data.total} SKU{data.total === 1 ? '' : 's'} at your warehouse
        </p>
      )}

      <div className="flex flex-wrap items-center gap-2 mb-4">
        {TABS.map((t) => (
          <button
            key={t.value}
            onClick={() => setStatusFilter(t.value)}
            className={`text-xs px-3 py-1.5 rounded-lg transition-colors ${
              statusFilter === t.value ? 'bg-indigo-500/20 text-indigo-300' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
            }`}
          >
            {t.label(data)}
          </button>
        ))}
        <form onSubmit={submitSearch} className="ml-auto flex gap-1.5">
          <input
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search product or barcode..."
            className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-slate-300 w-56 focus:outline-none focus:border-indigo-500"
          />
          <button type="submit" className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300">
            Search
          </button>
        </form>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading inventory...</p>}

      {!isLoading && rows.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No products found for this filter.
        </div>
      )}

      {!isLoading && rows.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Product</th>
                <th className="text-left px-4 py-2.5">Category</th>
                <th className="text-left px-4 py-2.5">Location</th>
                <th className="text-right px-4 py-2.5">Stock</th>
                <th className="text-right px-4 py-2.5">Reserved</th>
                <th className="text-right px-4 py-2.5">Available</th>
                <th className="text-left px-4 py-2.5">Status</th>
                <th className="text-left px-4 py-2.5">Flags</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {rows.map((r) => (
                <tr key={r.product_id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3">
                    <p className="font-medium">{r.product_name}</p>
                    {r.barcode && <p className="text-xs text-slate-500">{r.barcode}</p>}
                  </td>
                  <td className="px-4 py-3 text-slate-400">{r.category_name || '-'}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">
                    {r.bin_name ? `${r.zone_name} / ${r.rack_name} / ${r.bin_name}` : 'Unassigned'}
                  </td>
                  <td className="px-4 py-3 text-right">{r.stock}</td>
                  <td className="px-4 py-3 text-right text-slate-400">{r.reserved}</td>
                  <td className="px-4 py-3 text-right font-medium">{r.available}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`text-xs px-2 py-1 rounded-full ${
                        r.stock_status === 'out'
                          ? 'bg-rose-500/15 text-rose-300'
                          : r.stock_status === 'low'
                          ? 'bg-amber-500/15 text-amber-300'
                          : 'bg-emerald-500/15 text-emerald-300'
                      }`}
                    >
                      {r.stock_status === 'out' ? 'Out of Stock' : r.stock_status === 'low' ? 'Low Stock' : 'In Stock'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs space-y-1">
                    {r.expired_qty > 0 && (
                      <div className="text-rose-400">{r.expired_qty} units expired</div>
                    )}
                    {r.last_damaged_at && (
                      <div className="text-amber-400">
                        {r.last_damaged_qty} damaged &middot; {timeAgo(r.last_damaged_at)}
                      </div>
                    )}
                    {r.expired_qty === 0 && !r.last_damaged_at && <span className="text-slate-600">-</span>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!isLoading && data && data.total_pages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="text-xs text-slate-500">
            Page {page} of {data.total_pages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(data.total_pages, p + 1))}
            disabled={page >= data.total_pages}
            className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
