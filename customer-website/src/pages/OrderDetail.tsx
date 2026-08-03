import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { getOrder, requestReturn, getOrderTracking } from '../api/orders'
import { IMAGE_ORIGIN } from '../api/client'
import type { Order, OrderTracking } from '../types'

export default function OrderDetail() {
  const { id } = useParams<{ id: string }>()
  const [order, setOrder] = useState<Order | null>(null)
  const [showReturnForm, setShowReturnForm] = useState(false)
  const [returnReason, setReturnReason] = useState('')
  const [selectedItems, setSelectedItems] = useState<Record<number, number>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [tracking, setTracking] = useState<OrderTracking | null>(null)
  const [isLoadingTracking, setIsLoadingTracking] = useState(false)

  function load() {
    if (!id) return
    getOrder(parseInt(id, 10)).then(setOrder)
  }

  useEffect(() => {
    load()
  }, [id])

  useEffect(() => {
    if (!order || !id) return
    if (!['confirmed', 'shipped'].includes(order.status)) {
      setTracking(null)
      return
    }
    setIsLoadingTracking(true)
    getOrderTracking(parseInt(id, 10))
      .then(setTracking)
      .catch(() => setTracking(null))
      .finally(() => setIsLoadingTracking(false))
  }, [order?.status, id])

  function imageUrl(path: string) {
    if (!path) return ''
    return path.startsWith('http') ? path : `${IMAGE_ORIGIN}${path}`
  }

  async function handleReturnSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!order) return
    const items = Object.entries(selectedItems)
      .filter(([, qty]) => qty > 0)
      .map(([orderItemId, qty]) => ({ order_item_id: parseInt(orderItemId, 10), quantity: qty }))

    if (items.length === 0) {
      setMessage('Select at least one item to return.')
      return
    }

    setIsSubmitting(true)
    try {
      await requestReturn(order.id, { reason: returnReason, items })
      setMessage('Return request submitted.')
      setShowReturnForm(false)
      load()
    } catch (err: any) {
      setMessage(err.response?.data?.error ?? 'Failed to submit return request.')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!order) {
    return <div className="max-w-3xl mx-auto px-6 py-16 text-ink/50">Loading...</div>
  }

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      <p className="font-mono text-sm text-ink/50 mb-1">Order #{order.id}</p>
      <h1 className="font-display text-3xl font-600 mb-2 capitalize">{order.status}</h1>
      <p className="text-ink/60 mb-8">
        Placed on{' '}
        {new Date(order.created_at).toLocaleDateString('en-IN', {
          day: 'numeric',
          month: 'long',
          year: 'numeric',
        })}
      </p>

      <section className="border border-line rounded-xl divide-y divide-line mb-6">
        {order.items?.map((item) => (
          <div key={item.id} className="flex items-center gap-4 p-4">
            <div className="w-14 h-14 rounded-lg bg-line/30 overflow-hidden shrink-0">
              {item.product?.image_url ? (
                <img
                  src={imageUrl(item.product.image_url)}
                  alt={item.product.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center">📦</div>
              )}
            </div>
            <div className="flex-1">
              <p className="text-sm font-medium">{item.product?.name}</p>
              <p className="text-xs text-ink/50">
                Qty {item.quantity} · ₹{item.price} each
              </p>
            </div>
            <p className="font-mono text-sm font-semibold">₹{(item.price * item.quantity).toFixed(2)}</p>
          </div>
        ))}
      </section>

      {['confirmed', 'shipped'].includes(order.status) && (
        <section className="mb-6 border border-line rounded-xl p-4 bg-leaf/5">
          <h2 className="font-medium mb-2">Live tracking</h2>
          {isLoadingTracking ? (
            <p className="text-sm text-ink/50">Loading tracking...</p>
          ) : !tracking ? (
            <p className="text-sm text-ink/50">No delivery partner assigned yet.</p>
          ) : (
            <div className="text-sm space-y-1">
              <p>
                <span className="text-ink/60">Delivery partner:</span>{' '}
                <span className="font-medium">{tracking.delivery_partner_name}</span>
              </p>
              {tracking.vehicle_number && (
                <p>
                  <span className="text-ink/60">Vehicle:</span>{' '}
                  <span className="font-mono">{tracking.vehicle_number}</span>
                </p>
              )}
              {tracking.current_lat != null && tracking.current_lng != null ? (
                <p className="text-ink/60">
                  📍 Current location: {tracking.current_lat.toFixed(4)}, {tracking.current_lng.toFixed(4)}
                </p>
              ) : (
                <p className="text-ink/50">Location not available yet.</p>
              )}
              {tracking.last_updated && (
                <p className="text-xs text-ink/40">
                  Updated {new Date(tracking.last_updated).toLocaleString('en-IN')}
                </p>
              )}
            </div>
          )}
        </section>
      )}

      {order.address && (
        <section className="mb-6">
          <h2 className="font-medium mb-2">Delivered to</h2>
          <p className="text-sm text-ink/60">
            {order.address.full_name} · {order.address.phone}
            <br />
            {order.address.line1}, {order.address.city}, {order.address.state} - {order.address.pincode}
          </p>
        </section>
      )}

      <section className="border-t border-line pt-4 space-y-1.5 text-sm mb-8">
        <div className="flex justify-between">
          <span className="text-ink/60">Items</span>
          <span className="font-mono">₹{order.items_amount.toFixed(2)}</span>
        </div>
        {order.delivery_charge > 0 && (
          <div className="flex justify-between">
            <span className="text-ink/60">Delivery</span>
            <span className="font-mono">₹{order.delivery_charge.toFixed(2)}</span>
          </div>
        )}
        {order.wallet_amount_used > 0 && (
          <div className="flex justify-between text-leaf">
            <span>Wallet used</span>
            <span className="font-mono">−₹{order.wallet_amount_used.toFixed(2)}</span>
          </div>
        )}
        <div className="flex justify-between font-semibold text-base pt-2 border-t border-line">
          <span>Total</span>
          <span className="font-mono">₹{order.total_amount.toFixed(2)}</span>
        </div>
      </section>

      {order.status === 'delivered' && (
        <section>
          {!showReturnForm ? (
            <button
              onClick={() => setShowReturnForm(true)}
              className="text-sm font-medium text-clay hover:underline"
            >
              Request a return
            </button>
          ) : (
            <form onSubmit={handleReturnSubmit} className="border border-line rounded-xl p-5">
              <p className="text-sm font-medium mb-3">Select items to return</p>
              <div className="space-y-2 mb-4">
                {order.items?.map((item) => (
                  <label key={item.id} className="flex items-center gap-3 text-sm">
                    <input
                      type="checkbox"
                      onChange={(e) =>
                        setSelectedItems((prev) => ({
                          ...prev,
                          [item.id]: e.target.checked ? item.quantity : 0,
                        }))
                      }
                    />
                    {item.product?.name} (Qty {item.quantity})
                  </label>
                ))}
              </div>
              <textarea
                value={returnReason}
                onChange={(e) => setReturnReason(e.target.value)}
                placeholder="Reason for return"
                className="w-full border border-line rounded-lg px-3 py-2 outline-none focus:border-ink mb-3 resize-none"
                rows={3}
                required
              />
              {message && <p className="text-sm text-clay mb-3">{message}</p>}
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="bg-ink text-paper text-sm font-medium px-4 py-2 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
                >
                  {isSubmitting ? 'Submitting...' : 'Submit return request'}
                </button>
                <button
                  type="button"
                  onClick={() => setShowReturnForm(false)}
                  className="text-sm px-4 py-2 rounded-lg border border-line hover:border-ink"
                >
                  Cancel
                </button>
              </div>
            </form>
          )}
        </section>
      )}

      {message && !showReturnForm && <p className="text-sm text-leaf mt-4">{message}</p>}
    </div>
  )
}
