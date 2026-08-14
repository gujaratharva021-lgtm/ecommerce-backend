import { useEffect, useState, useCallback } from 'react'
import { listBatches, listExpiringBatches, createBatch, adjustBatchQuantity, deleteBatch } from '../api/warehouse'
import type { Batch } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

function daysUntil(dateStr: string): number {
  const ms = new Date(dateStr).getTime() - Date.now()
  return Math.ceil(ms / (1000 * 60 * 60 * 24))
}

function ExpiryBadge({ expiryDate }: { expiryDate: string }) {
  const days = daysUntil(expiryDate)
  let style = 'bg-slate-700 text-slate-300'
  let label = `${days} day${days === 1 ? '' : 's'} left`
  if (days < 0) {
    style = 'bg-rose-500/15 text-rose-300'
    label = `Expired ${Math.abs(days)}d ago`
  } else if (days <= 3) {
    style = 'bg-rose-500/15 text-rose-300'
  } else if (days <= 7) {
    style = 'bg-amber-500/15 text-amber-300'
  } else {
    style = 'bg-emerald-500/15 text-emerald-300'
  }
  return <span className={`inline-block px-2 py-0.5 rounded-md text-xs font-medium ${style}`}>{label}</span>
}

export default function Batches() {
  const [batches, setBatches] = useState<Batch[]>([])
  const [expiringCount, setExpiringCount] = useState<number | null>(null)
  const [expiringWindow, setExpiringWindow] = useState(7)
  const [showExpiringOnly, setShowExpiringOnly] = useState(false)
  const [productIdFilter, setProductIdFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showCreate, setShowCreate] = useState(false)
  const [newProductId, setNewProductId] = useState('')
  const [newBatchNumber, setNewBatchNumber] = useState('')
  const [newManufactureDate, setNewManufactureDate] = useState('')
  const [newExpiryDate, setNewExpiryDate] = useState('')
  const [newQuantity, setNewQuantity] = useState('')
  const [newBinId, setNewBinId] = useState('')
  const [isCreating, setIsCreating] = useState(false)

  const [selected, setSelected] = useState<Batch | null>(null)
  const [adjustQty, setAdjustQty] = useState('')
  const [adjustReason, setAdjustReason] = useState('')
  const [isAdjusting, setIsAdjusting] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    setIsLoading(true)
    try {
      if (showExpiringOnly) {
        const data = await listExpiringBatches(expiringWindow)
        setBatches(data.batches)
        setTotalPages(1)
      } else {
        const data = await listBatches({
          product_id: productIdFilter ? parseInt(productIdFilter, 10) : undefined,
          page,
          limit: 20,
        })
        setBatches(data.batches)
        setTotalPages(data.total_pages || 1)
      }
      const expiring = await listExpiringBatches(7)
      setExpiringCount(expiring.count)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load batches.'))
    } finally {
      setIsLoading(false)
    }
  }, [showExpiringOnly, expiringWindow, productIdFilter, page])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    setPage(1)
  }, [productIdFilter, showExpiringOnly])

  const resetCreateForm = () => {
    setNewProductId('')
    setNewBatchNumber('')
    setNewManufactureDate('')
    setNewExpiryDate('')
    setNewQuantity('')
    setNewBinId('')
  }

  const submitCreate = async () => {
    if (!newProductId || !newBatchNumber.trim() || !newExpiryDate || !newQuantity) return
    setIsCreating(true)
    setError(null)
    try {
      await createBatch({
        product_id: parseInt(newProductId, 10),
        batch_number: newBatchNumber.trim(),
        manufacture_date: newManufactureDate ? new Date(newManufactureDate).toISOString() : undefined,
        expiry_date: new Date(newExpiryDate).toISOString(),
        quantity: parseInt(newQuantity, 10),
        bin_id: newBinId ? parseInt(newBinId, 10) : undefined,
      })
      resetCreateForm()
      setShowCreate(false)
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create batch.'))
    } finally {
      setIsCreating(false)
    }
  }

  const openAdjust = (batch: Batch) => {
    setSelected(batch)
    setAdjustQty(String(batch.quantity))
    setAdjustReason('')
    setActionError(null)
  }

  const submitAdjust = async () => {
    if (!selected) return
    const qty = parseInt(adjustQty, 10)
    if (isNaN(qty) || qty < 0) {
      setActionError('Enter a valid non-negative quantity.')
      return
    }
    if (!adjustReason.trim()) {
      setActionError('A reason is required.')
      return
    }
    setIsAdjusting(true)
    setActionError(null)
    try {
      const updated = await adjustBatchQuantity(selected.id, { quantity: qty, reason: adjustReason.trim() })
      setSelected(updated)
      await load()
    } catch (err) {
      setActionError(getErrorMessage(err, 'Failed to adjust batch quantity.'))
    } finally {
      setIsAdjusting(false)
    }
  }

  const submitDelete = async () => {
    if (!selected) return
    if (selected.quantity > 0) {
      setActionError('Adjust quantity to 0 before deleting this batch.')
      return
    }
    if (!confirm(`Delete batch "${selected.batch_number}"? This cannot be undone.`)) return
    setIsDeleting(true)
    setActionError(null)
    try {
      await deleteBatch(selected.id)
      setSelected(null)
      await load()
    } catch (err) {
      setActionError(getErrorMessage(err, 'Failed to delete batch.'))
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-2">
        <h1 className="font-display text-2xl font-semibold">Batch & Expiry</h1>
        <div className="flex gap-2">
          <button onClick={load} className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors">
            Refresh
          </button>
          <button
            onClick={() => setShowCreate(true)}
            className="text-xs px-3 py-1.5 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 transition-colors"
          >
            + New Batch
          </button>
        </div>
      </div>
      <p className="text-xs text-slate-500 mb-6">
        Listing is FEFO-ordered (earliest expiry first) — the order shown here is the pick priority.
      </p>

      {expiringCount !== null && expiringCount > 0 && !showExpiringOnly && (
        <button
          onClick={() => setShowExpiringOnly(true)}
          className="w-full text-left text-sm px-4 py-2.5 rounded-lg border border-amber-900 bg-amber-950/30 text-amber-300 mb-4 hover:brightness-125"
        >
          {expiringCount} batch{expiringCount === 1 ? '' : 'es'} expiring within 7 days — click to filter
        </button>
      )}

      <div className="flex flex-wrap items-center gap-2 mb-4">
        <button
          onClick={() => setShowExpiringOnly(false)}
          className={`text-xs px-3 py-1.5 rounded-lg transition-colors ${
            !showExpiringOnly ? 'bg-indigo-500/20 text-indigo-300' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
          }`}
        >
          All Batches
        </button>
        <button
          onClick={() => setShowExpiringOnly(true)}
          className={`text-xs px-3 py-1.5 rounded-lg transition-colors ${
            showExpiringOnly ? 'bg-indigo-500/20 text-indigo-300' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
          }`}
        >
          Expiring Soon
        </button>
        {showExpiringOnly && (
          <select
            value={expiringWindow}
            onChange={(e) => setExpiringWindow(Number(e.target.value))}
            className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-slate-300"
          >
            <option value={3}>Within 3 days</option>
            <option value={7}>Within 7 days</option>
            <option value={14}>Within 14 days</option>
            <option value={30}>Within 30 days</option>
          </select>
        )}
        {!showExpiringOnly && (
          <input
            value={productIdFilter}
            onChange={(e) => setProductIdFilter(e.target.value)}
            placeholder="Filter by product ID"
            className="text-xs bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-slate-300 w-40"
          />
        )}
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading batches...</p>}

      {!isLoading && batches.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No batches found for this filter.
        </div>
      )}

      {!isLoading && batches.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Product</th>
                <th className="text-left px-4 py-2.5">Batch #</th>
                <th className="text-left px-4 py-2.5">Bin</th>
                <th className="text-right px-4 py-2.5">Quantity</th>
                <th className="text-left px-4 py-2.5">Expiry</th>
                <th className="text-right px-4 py-2.5">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {batches.map((b) => (
                <tr key={b.id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3">{b.product?.name ?? `#${b.product_id}`}</td>
                  <td className="px-4 py-3 text-slate-300">{b.batch_number}</td>
                  <td className="px-4 py-3 text-slate-400">{b.bin?.name ?? '-'}</td>
                  <td className="px-4 py-3 text-right">{b.quantity}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-col gap-1">
                      <span className="text-slate-400 text-xs">{new Date(b.expiry_date).toLocaleDateString()}</span>
                      <ExpiryBadge expiryDate={b.expiry_date} />
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openAdjust(b)}
                      className="text-xs px-2.5 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-300"
                    >
                      Manage
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!isLoading && !showExpiringOnly && totalPages > 1 && (
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

      {/* Create modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 w-full max-w-md">
            <h2 className="text-sm font-semibold mb-4">New Batch</h2>
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="text-xs text-slate-500 block mb-1">Product ID</label>
                  <input
                    value={newProductId}
                    onChange={(e) => setNewProductId(e.target.value)}
                    className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-500 block mb-1">Batch number</label>
                  <input
                    value={newBatchNumber}
                    onChange={(e) => setNewBatchNumber(e.target.value)}
                    placeholder="LOT-2026-001"
                    className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="text-xs text-slate-500 block mb-1">Manufacture date (optional)</label>
                  <input
                    type="date"
                    value={newManufactureDate}
                    onChange={(e) => setNewManufactureDate(e.target.value)}
                    className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-500 block mb-1">Expiry date</label>
                  <input
                    type="date"
                    value={newExpiryDate}
                    onChange={(e) => setNewExpiryDate(e.target.value)}
                    className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="text-xs text-slate-500 block mb-1">Quantity</label>
                  <input
                    type="number"
                    min={1}
                    value={newQuantity}
                    onChange={(e) => setNewQuantity(e.target.value)}
                    className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-500 block mb-1">Bin ID (optional)</label>
                  <input
                    value={newBinId}
                    onChange={(e) => setNewBinId(e.target.value)}
                    className="w-full text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>
            </div>
            <div className="flex gap-2 justify-end mt-4">
              <button onClick={() => setShowCreate(false)} disabled={isCreating} className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700">
                Cancel
              </button>
              <button
                onClick={submitCreate}
                disabled={isCreating || !newProductId || !newBatchNumber.trim() || !newExpiryDate || !newQuantity}
                className="text-xs px-3 py-1.5 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-50"
              >
                {isCreating ? 'Creating...' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Manage modal */}
      {selected && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 w-full max-w-md">
            <div className="flex items-center justify-between mb-1">
              <h2 className="text-sm font-semibold">Batch {selected.batch_number}</h2>
              <ExpiryBadge expiryDate={selected.expiry_date} />
            </div>
            <p className="text-xs text-slate-500 mb-4">
              {selected.product?.name ?? `Product #${selected.product_id}`} · Current quantity:{' '}
              <span className="text-slate-300">{selected.quantity}</span>
            </p>

            {actionError && (
              <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-xs rounded-lg px-3 py-2 mb-3">{actionError}</div>
            )}

            <div className="space-y-3 mb-4">
              <h3 className="text-xs font-medium text-slate-300">Adjust Quantity</h3>
              <div className="grid grid-cols-2 gap-2">
                <input
                  type="number"
                  min={0}
                  value={adjustQty}
                  onChange={(e) => setAdjustQty(e.target.value)}
                  className="text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                />
                <input
                  value={adjustReason}
                  onChange={(e) => setAdjustReason(e.target.value)}
                  placeholder="Reason (e.g. FEFO pick, recount)"
                  className="text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                />
              </div>
              <button
                onClick={submitAdjust}
                disabled={isAdjusting}
                className="w-full text-xs px-3 py-2 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-50"
              >
                {isAdjusting ? 'Saving...' : 'Save Quantity'}
              </button>
            </div>

            <div className="border-t border-slate-800 pt-3">
              <button
                onClick={submitDelete}
                disabled={isDeleting || selected.quantity > 0}
                title={selected.quantity > 0 ? 'Quantity must be 0 before deleting' : undefined}
                className="w-full text-xs px-3 py-2 rounded-lg bg-rose-600/80 hover:bg-rose-600 text-white disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {isDeleting ? 'Deleting...' : 'Delete Batch (quantity must be 0)'}
              </button>
            </div>

            <button
              onClick={() => setSelected(null)}
              className="w-full text-xs px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 mt-3"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
