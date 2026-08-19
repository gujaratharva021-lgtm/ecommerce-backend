import { useCallback, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  listSubstitutionRequests,
  createSubstitutionRequest,
  approveSubstitutionRequest,
  rejectSubstitutionRequest,
  listProducts,
} from '../api/warehouse'
import type { SubstitutionRequest, Product } from '../types/warehouse'
import StatusBadge from '../components/StatusBadge'
import { useAuth } from '../context/AuthContext'
import { getErrorMessage } from '../utils/errors'

const TABS: { value: string; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
]

function NewRequestModal({ defaultOrderId, defaultItemId, defaultProductId, onClose, onCreated }: {
  defaultOrderId?: string | null
  defaultItemId?: string | null
  defaultProductId?: string | null
  onClose: () => void
  onCreated: () => void
}) {
  const [orderId, setOrderId] = useState(defaultOrderId ?? '')
  const [originalProductId, setOriginalProductId] = useState(defaultProductId ?? '')
  const [substituteQuery, setSubstituteQuery] = useState('')
  const [substituteOptions, setSubstituteOptions] = useState<Product[]>([])
  const [substituteProductId, setSubstituteProductId] = useState('')
  const [quantity, setQuantity] = useState('1')
  const [reason, setReason] = useState('')
  const [isSearching, setIsSearching] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!substituteQuery.trim()) {
      setSubstituteOptions([])
      return
    }
    const t = setTimeout(async () => {
      setIsSearching(true)
      try {
        const res = await listProducts({ search: substituteQuery, limit: 8 })
        setSubstituteOptions(res.products ?? res.data ?? [])
      } catch {
        setSubstituteOptions([])
      } finally {
        setIsSearching(false)
      }
    }, 300)
    return () => clearTimeout(t)
  }, [substituteQuery])

  async function handleSubmit() {
    if (!orderId || !originalProductId || !substituteProductId || !quantity) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createSubstitutionRequest({
        order_id: Number(orderId),
        picking_task_item_id: defaultItemId ? Number(defaultItemId) : undefined,
        original_product_id: Number(originalProductId),
        substitute_product_id: Number(substituteProductId),
        quantity: Number(quantity),
        reason: reason || undefined,
      })
      onCreated()
      onClose()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create substitution request.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 w-full max-w-md">
        <h2 className="text-base font-semibold mb-4">Request Substitution</h2>

        {error && (
          <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-xs rounded-lg px-3 py-2 mb-3">
            {error}
          </div>
        )}

        <div className="space-y-3">
          <div>
            <label className="block text-xs text-slate-400 mb-1">Order ID</label>
            <input
              type="number"
              value={orderId}
              onChange={(e) => setOrderId(e.target.value)}
              disabled={!!defaultOrderId}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs disabled:opacity-60"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-400 mb-1">Original Product ID</label>
            <input
              type="number"
              value={originalProductId}
              onChange={(e) => setOriginalProductId(e.target.value)}
              disabled={!!defaultProductId}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs disabled:opacity-60"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-400 mb-1">Substitute Product</label>
            <input
              type="text"
              placeholder="Search product by name..."
              value={substituteQuery}
              onChange={(e) => setSubstituteQuery(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
            />
            {isSearching && <p className="text-xs text-slate-500 mt-1">Searching...</p>}
            {substituteOptions.length > 0 && (
              <div className="mt-1 border border-slate-800 rounded-lg overflow-hidden max-h-40 overflow-y-auto">
                {substituteOptions.map((p) => (
                  <button
                    key={p.id}
                    onClick={() => {
                      setSubstituteProductId(String(p.id))
                      setSubstituteQuery(p.name)
                      setSubstituteOptions([])
                    }}
                    className={`w-full text-left px-3 py-1.5 text-xs hover:bg-slate-800 ${
                      substituteProductId === String(p.id) ? 'bg-indigo-500/10 text-indigo-300' : ''
                    }`}
                  >
                    {p.name} (#{p.id})
                  </button>
                ))}
              </div>
            )}
          </div>

          <div>
            <label className="block text-xs text-slate-400 mb-1">Quantity</label>
            <input
              type="number"
              min={1}
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-400 mb-1">Reason (optional)</label>
            <input
              type="text"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="e.g. out of stock"
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
            />
          </div>
        </div>

        <div className="flex justify-end gap-2 mt-5">
          <button
            onClick={onClose}
            className="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200"
          >
            Cancel
          </button>
          <button
            disabled={isSubmitting || !orderId || !originalProductId || !substituteProductId || !quantity}
            onClick={handleSubmit}
            className="px-4 py-1.5 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-xs font-medium disabled:opacity-50"
          >
            {isSubmitting ? 'Submitting...' : 'Submit Request'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default function Substitution() {
  const { staff } = useAuth()
  const isManager = staff?.role === 'warehouse_manager'
  const [searchParams] = useSearchParams()
  const [statusFilter, setStatusFilter] = useState('')
  const [requests, setRequests] = useState<SubstitutionRequest[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showModal, setShowModal] = useState(false)
  const [decidingId, setDecidingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listSubstitutionRequests({ status: statusFilter || undefined, limit: 50 })
      setRequests(res.substitution_requests ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load substitution requests.'))
    } finally {
      setIsLoading(false)
    }
  }, [statusFilter])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (searchParams.get('order_id')) setShowModal(true)
  }, [searchParams])

  async function handleDecision(id: number, action: 'approve' | 'reject') {
    setDecidingId(id)
    setError(null)
    try {
      if (action === 'approve') await approveSubstitutionRequest(id)
      else await rejectSubstitutionRequest(id)
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to update substitution request.'))
    } finally {
      setDecidingId(null)
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-1">
        <h1 className="font-display text-2xl font-semibold">Substitution Requests</h1>
        <button
          onClick={() => setShowModal(true)}
          className="px-3 py-1.5 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-xs font-medium"
        >
          + New Request
        </button>
      </div>
      <p className="text-xs text-slate-500 mb-5">
        {isManager
          ? 'Approve or reject substitution requests from pickers and packers.'
          : 'Request a product swap when the original item is unavailable or short.'}
      </p>

      <div className="flex gap-1 mb-4">
        {TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setStatusFilter(tab.value)}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
              statusFilter === tab.value
                ? 'bg-indigo-500/20 text-indigo-300'
                : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="text-sm text-slate-400">Loading...</div>
      ) : requests.length === 0 ? (
        <div className="text-sm text-slate-500 border border-slate-800 rounded-xl p-6 text-center">
          No substitution requests found.
        </div>
      ) : (
        <div className="space-y-3">
          {requests.map((req) => (
            <div key={req.id} className="border border-slate-800 rounded-xl bg-slate-900 p-4">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-xs text-slate-500 mb-1">
                    Order #{req.order_id} &middot; Requested {new Date(req.created_at).toLocaleString()}
                  </p>
                  <p className="text-sm">
                    <span className="text-slate-400">{req.original_product?.name ?? `Product #${req.original_product_id}`}</span>
                    {' '}&rarr;{' '}
                    <span className="font-medium text-slate-100">
                      {req.substitute_product?.name ?? `Product #${req.substitute_product_id}`}
                    </span>
                    <span className="text-slate-500"> &times;{req.quantity}</span>
                  </p>
                  {req.reason && <p className="text-xs text-amber-400 mt-1">Reason: {req.reason}</p>}
                  {req.decision_note && (
                    <p className="text-xs text-slate-500 mt-1">Note: {req.decision_note}</p>
                  )}
                </div>
                <StatusBadge status={req.status} />
              </div>

              {isManager && req.status === 'pending' && (
                <div className="flex gap-2 mt-3">
                  <button
                    disabled={decidingId === req.id}
                    onClick={() => handleDecision(req.id, 'approve')}
                    className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium disabled:opacity-50"
                  >
                    Approve
                  </button>
                  <button
                    disabled={decidingId === req.id}
                    onClick={() => handleDecision(req.id, 'reject')}
                    className="px-3 py-1.5 rounded-lg bg-rose-600/80 hover:bg-rose-600 text-white text-xs font-medium disabled:opacity-50"
                  >
                    Reject
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showModal && (
        <NewRequestModal
          defaultOrderId={searchParams.get('order_id')}
          defaultItemId={searchParams.get('item_id')}
          defaultProductId={searchParams.get('product_id')}
          onClose={() => setShowModal(false)}
          onCreated={load}
        />
      )}
    </div>
  )
}
