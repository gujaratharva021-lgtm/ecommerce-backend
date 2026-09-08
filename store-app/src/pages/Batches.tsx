import { useCallback, useEffect, useState } from 'react'
import { listBatches, listExpiringBatches, createBatch, deleteBatch, listProducts } from '../api/warehouse'
import type { Batch, Product } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

function daysUntil(dateStr: string) {
  const diff = new Date(dateStr).getTime() - Date.now()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

function NewBatchModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [query, setQuery] = useState('')
  const [options, setOptions] = useState<Product[]>([])
  const [productId, setProductId] = useState('')
  const [batchNumber, setBatchNumber] = useState('')
  const [manufactureDate, setManufactureDate] = useState('')
  const [expiryDate, setExpiryDate] = useState('')
  const [quantity, setQuantity] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!query.trim()) { setOptions([]); return }
    const t = setTimeout(() => {
      listProducts({ search: query, limit: 10 }).then((res) => setOptions(res.products ?? res ?? [])).catch(() => {})
    }, 300)
    return () => clearTimeout(t)
  }, [query])

  async function handleSubmit() {
    if (!productId || !batchNumber.trim() || !expiryDate || !quantity) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createBatch({
        product_id: Number(productId),
        batch_number: batchNumber.trim(),
        manufacture_date: manufactureDate || undefined,
        expiry_date: expiryDate,
        quantity: Number(quantity),
      })
      onCreated()
      onClose()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create batch.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 w-full max-w-md">
        <h3 className="text-base font-semibold text-slate-100 mb-4">New Batch</h3>
        {error && <p className="text-xs text-red-400 mb-3">{error}</p>}
        <div className="space-y-3">
          <div>
            <label className="block text-xs text-slate-400 mb-1">Product</label>
            <input
              type="text"
              value={query}
              onChange={(e) => { setQuery(e.target.value); setProductId('') }}
              placeholder="Search product..."
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
            />
            {options.length > 0 && !productId && (
              <div className="mt-1 border border-slate-700 rounded-lg overflow-hidden max-h-32 overflow-y-auto">
                {options.map((p) => (
                  <div
                    key={p.id}
                    onClick={() => { setProductId(String(p.id)); setQuery(p.name); setOptions([]) }}
                    className="px-3 py-1.5 text-xs hover:bg-slate-800 cursor-pointer"
                  >
                    {p.name}
                  </div>
                ))}
              </div>
            )}
          </div>
          <div>
            <label className="block text-xs text-slate-400 mb-1">Batch number</label>
            <input
              type="text"
              value={batchNumber}
              onChange={(e) => setBatchNumber(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-slate-400 mb-1">Manufacture date</label>
              <input
                type="date"
                value={manufactureDate}
                onChange={(e) => setManufactureDate(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1">Expiry date</label>
              <input
                type="date"
                value={expiryDate}
                onChange={(e) => setExpiryDate(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
              />
            </div>
          </div>
          <div>
            <label className="block text-xs text-slate-400 mb-1">Quantity</label>
            <input
              type="number"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
            />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-5">
          <button onClick={onClose} className="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200">
            Cancel
          </button>
          <button
            disabled={isSubmitting || !productId || !batchNumber.trim() || !expiryDate || !quantity}
            onClick={handleSubmit}
            className="px-4 py-1.5 rounded-lg bg-red-500 hover:bg-red-400 text-white text-xs font-medium disabled:opacity-50"
          >
            {isSubmitting ? 'Creating...' : 'Create Batch'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default function Batches() {
  const [tab, setTab] = useState<'all' | 'expiring'>('all')
  const [batches, setBatches] = useState<Batch[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showModal, setShowModal] = useState(false)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      if (tab === 'expiring') {
        const res = await listExpiringBatches(30)
        setBatches(res.batches ?? [])
      } else {
        const res = await listBatches({ limit: 100 })
        setBatches(res.batches ?? [])
      }
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load batches.'))
    } finally {
      setIsLoading(false)
    }
  }, [tab])

  useEffect(() => { load() }, [load])

  async function handleDelete(id: number) {
    setError(null)
    try {
      await deleteBatch(id)
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete batch.'))
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-1">
        <h1 className="font-display text-2xl font-semibold">Batches</h1>
        <button
          onClick={() => setShowModal(true)}
          className="px-4 py-1.5 rounded-lg bg-red-500 hover:bg-red-400 text-white text-xs font-medium"
        >
          New Batch
        </button>
      </div>
      <p className="text-sm text-slate-400 mb-5">Track batch numbers and expiry dates.</p>

      <div className="flex gap-2 mb-4">
        {(['all', 'expiring'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium ${
              tab === t ? 'bg-red-500 text-white' : 'bg-slate-800 text-slate-400 hover:text-slate-200'
            }`}
          >
            {t === 'all' ? 'All Batches' : 'Near Expiry (30d)'}
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
      ) : batches.length === 0 ? (
        <p className="text-sm text-slate-500">No batches found.</p>
      ) : (
        <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs text-slate-500 border-b border-slate-800">
                <th className="px-4 py-2.5 font-medium">Product</th>
                <th className="px-4 py-2.5 font-medium">Batch #</th>
                <th className="px-4 py-2.5 font-medium">Expiry</th>
                <th className="px-4 py-2.5 font-medium">Quantity</th>
                <th className="px-4 py-2.5 font-medium">Bin</th>
                <th className="px-4 py-2.5 font-medium"></th>
              </tr>
            </thead>
            <tbody>
              {batches.map((b) => {
                const days = daysUntil(b.expiry_date)
                return (
                  <tr key={b.id} className="border-b border-slate-800/60 last:border-0">
                    <td className="px-4 py-2.5 text-slate-200">{b.product?.name ?? `#${b.product_id}`}</td>
                    <td className="px-4 py-2.5 text-slate-400 font-mono text-xs">{b.batch_number}</td>
                    <td className="px-4 py-2.5">
                      <span className={days <= 7 ? 'text-red-400 font-medium' : days <= 30 ? 'text-amber-400' : 'text-slate-300'}>
                        {b.expiry_date} {days >= 0 ? `(${days}d)` : '(expired)'}
                      </span>
                    </td>
                    <td className="px-4 py-2.5 text-slate-300">{b.quantity}</td>
                    <td className="px-4 py-2.5 text-slate-400">{b.bin?.name ?? '-'}</td>
                    <td className="px-4 py-2.5 text-right">
                      <button onClick={() => handleDelete(b.id)} className="text-xs text-slate-500 hover:text-red-400">
                        Delete
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {showModal && <NewBatchModal onClose={() => setShowModal(false)} onCreated={load} />}
    </div>
  )
}

