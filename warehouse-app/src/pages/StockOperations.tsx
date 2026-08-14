import { useEffect, useState, useCallback } from 'react'
import { getProductInventory, adjustStock, listStockMovements } from '../api/warehouse'
import type { Inventory, StockMovement, AdjustReason } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

const REASONS: { value: AdjustReason; label: string }[] = [
  { value: 'damaged', label: 'Damaged' },
  { value: 'expired', label: 'Expired' },
  { value: 'counting_error', label: 'Counting error' },
  { value: 'lost', label: 'Lost' },
  { value: 'found', label: 'Found' },
  { value: 'manual_correction', label: 'Manual correction' },
  { value: 'other', label: 'Other' },
]

const MOVEMENT_TONES: Record<string, string> = {
  receive: 'text-emerald-300',
  sale: 'text-sky-300',
  adjustment: 'text-amber-300',
  transfer: 'text-violet-300',
  damaged: 'text-rose-300',
  expired: 'text-rose-300',
  return: 'text-emerald-300',
  correction: 'text-amber-300',
}

export default function StockOperations() {
  // Adjustment form
  const [productIdInput, setProductIdInput] = useState('')
  const [inventory, setInventory] = useState<Inventory | null>(null)
  const [isLookingUp, setIsLookingUp] = useState(false)
  const [lookupError, setLookupError] = useState<string | null>(null)

  const [newQuantity, setNewQuantity] = useState('')
  const [reason, setReason] = useState<AdjustReason>('counting_error')
  const [notes, setNotes] = useState('')
  const [isAdjusting, setIsAdjusting] = useState(false)
  const [adjustSuccess, setAdjustSuccess] = useState(false)

  // Movement history
  const [movements, setMovements] = useState<StockMovement[]>([])
  const [movementTypeFilter, setMovementTypeFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoadingMovements, setIsLoadingMovements] = useState(true)
  const [movementsError, setMovementsError] = useState<string | null>(null)

  const loadMovements = useCallback(async () => {
    setMovementsError(null)
    setIsLoadingMovements(true)
    try {
      const data = await listStockMovements({
        movement_type: movementTypeFilter || undefined,
        page,
        limit: 20,
      })
      setMovements(data.movements)
      setTotalPages(data.total_pages || 1)
    } catch (err) {
      setMovementsError(getErrorMessage(err, 'Failed to load stock movements.'))
    } finally {
      setIsLoadingMovements(false)
    }
  }, [movementTypeFilter, page])

  useEffect(() => {
    loadMovements()
  }, [loadMovements])

  useEffect(() => {
    setPage(1)
  }, [movementTypeFilter])

  const lookupProduct = async () => {
    const id = parseInt(productIdInput, 10)
    if (!id) return
    setIsLookingUp(true)
    setLookupError(null)
    setInventory(null)
    setAdjustSuccess(false)
    try {
      const inv = await getProductInventory(id)
      setInventory(inv)
      setNewQuantity(String(inv.stock))
    } catch (err) {
      setLookupError(getErrorMessage(err, 'No inventory found for this product in your warehouse.'))
    } finally {
      setIsLookingUp(false)
    }
  }

  const submitAdjustment = async () => {
    if (!inventory) return
    const qty = parseInt(newQuantity, 10)
    if (isNaN(qty) || qty < 0) {
      setLookupError('Enter a valid non-negative quantity.')
      return
    }
    setIsAdjusting(true)
    setLookupError(null)
    setAdjustSuccess(false)
    try {
      const updated = await adjustStock(inventory.product_id, {
        new_quantity: qty,
        reason,
        notes: notes.trim() || undefined,
      })
      setInventory(updated)
      setNotes('')
      setAdjustSuccess(true)
      await loadMovements()
    } catch (err) {
      setLookupError(getErrorMessage(err, 'Failed to adjust stock.'))
    } finally {
      setIsAdjusting(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl">
      <h1 className="font-display text-2xl font-semibold mb-6">Stock Operations</h1>

      {/* Adjustment form */}
      <div className="border border-slate-800 rounded-xl bg-slate-900 p-5 max-w-2xl mb-8">
        <h2 className="text-sm font-semibold mb-3">Stock Adjustment</h2>
        <div className="flex gap-2 mb-3">
          <input
            value={productIdInput}
            onChange={(e) => setProductIdInput(e.target.value)}
            placeholder="Product ID"
            className="flex-1 text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
          />
          <button
            onClick={lookupProduct}
            disabled={isLookingUp || !productIdInput}
            className="text-xs px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40"
          >
            {isLookingUp ? 'Looking up...' : 'Look up'}
          </button>
        </div>

        {lookupError && (
          <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-xs rounded-lg px-3 py-2 mb-3">
            {lookupError}
          </div>
        )}

        {adjustSuccess && (
          <div className="border border-emerald-900 bg-emerald-950/30 text-emerald-300 text-xs rounded-lg px-3 py-2 mb-3">
            Stock adjusted successfully.
          </div>
        )}

        {inventory && (
          <div className="border border-slate-800 rounded-lg p-3 bg-slate-950/50 space-y-3">
            <p className="text-sm font-medium">
              {inventory.product?.name ?? `Product #${inventory.product_id}`}
              <span className="text-slate-500 font-normal ml-2 text-xs">Current stock: {inventory.stock}</span>
            </p>

            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-xs text-slate-500 block mb-1">New quantity</label>
                <input
                  type="number"
                  min={0}
                  value={newQuantity}
                  onChange={(e) => setNewQuantity(e.target.value)}
                  className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                />
              </div>
              <div>
                <label className="text-xs text-slate-500 block mb-1">Reason</label>
                <select
                  value={reason}
                  onChange={(e) => setReason(e.target.value as AdjustReason)}
                  className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                >
                  {REASONS.map((r) => (
                    <option key={r.value} value={r.value}>
                      {r.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className="text-xs text-slate-500 block mb-1">Notes (optional)</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={2}
                className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>

            <button
              onClick={submitAdjustment}
              disabled={isAdjusting || parseInt(newQuantity, 10) === inventory.stock}
              className="text-xs px-4 py-2 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-40"
            >
              {isAdjusting ? 'Saving...' : 'Save Adjustment'}
            </button>
          </div>
        )}
      </div>

      {/* Movement history */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold">Stock Movement History</h2>
        <select
          value={movementTypeFilter}
          onChange={(e) => setMovementTypeFilter(e.target.value)}
          className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-slate-300"
        >
          <option value="">All types</option>
          <option value="receive">Receive</option>
          <option value="sale">Sale</option>
          <option value="adjustment">Adjustment</option>
          <option value="transfer">Transfer</option>
          <option value="damaged">Damaged</option>
          <option value="expired">Expired</option>
          <option value="return">Return</option>
          <option value="correction">Correction</option>
        </select>
      </div>

      {movementsError && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {movementsError}
        </div>
      )}

      {isLoadingMovements && <p className="text-sm text-slate-400">Loading movements...</p>}

      {!isLoadingMovements && movements.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No stock movements found.
        </div>
      )}

      {!isLoadingMovements && movements.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Product</th>
                <th className="text-left px-4 py-2.5">Type</th>
                <th className="text-right px-4 py-2.5">Previous</th>
                <th className="text-right px-4 py-2.5">Change</th>
                <th className="text-right px-4 py-2.5">New</th>
                <th className="text-left px-4 py-2.5">Reason</th>
                <th className="text-left px-4 py-2.5">Reference</th>
                <th className="text-left px-4 py-2.5">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {movements.map((m) => (
                <tr key={m.id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3">{m.product?.name ?? `#${m.product_id}`}</td>
                  <td className={`px-4 py-3 font-medium capitalize ${MOVEMENT_TONES[m.movement_type] ?? 'text-slate-300'}`}>
                    {m.movement_type}
                  </td>
                  <td className="px-4 py-3 text-right text-slate-400">{m.previous_qty}</td>
                  <td className={`px-4 py-3 text-right font-medium ${m.change >= 0 ? 'text-emerald-300' : 'text-rose-300'}`}>
                    {m.change >= 0 ? '+' : ''}
                    {m.change}
                  </td>
                  <td className="px-4 py-3 text-right">{m.new_qty}</td>
                  <td className="px-4 py-3 text-slate-400">{m.reason || '-'}</td>
                  <td className="px-4 py-3 text-slate-500 text-xs">
                    {m.reference_id ? `Order #${m.reference_id}` : '-'}
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{new Date(m.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!isLoadingMovements && totalPages > 1 && (
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
