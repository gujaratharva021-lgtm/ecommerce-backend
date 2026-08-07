import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listOrders, cancelOrder } from '../api/orders'
import type { Order } from '../types'

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-amber-100 text-amber-700',
  confirmed: 'bg-blue-100 text-blue-700',
  shipped: 'bg-indigo-100 text-indigo-700',
  delivered: 'bg-leaf/15 text-leaf',
  returned: 'bg-line text-ink/60',
  cancelled: 'bg-clay/15 text-clay',
}

export default function Orders() {
  const [orders, setOrders] = useState<Order[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [cancellingId, setCancellingId] = useState<number | null>(null)

  function load() {
    setIsLoading(true)
    listOrders()
      .then((res) => setOrders(res.orders ?? []))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCancel(id: number) {
    setCancellingId(id)
    try {
      await cancelOrder(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to cancel order.')
    } finally {
      setCancellingId(null)
    }
  }

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      <h1 className="font-display text-3xl font-600 mb-8">Your orders</h1>

      {isLoading ? (
        <p className="text-ink/50">Loading...</p>
      ) : orders.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-ink/60 mb-6">No orders yet.</p>
          <Link to="/products" className="text-marigold font-medium hover:underline">
            Start shopping →
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((o) => (
            <div key={o.id} className="border border-line rounded-xl p-5">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <p className="font-mono font-medium">Order #{o.id}</p>
                  <p className="text-xs text-ink/50">
                    {new Date(o.created_at).toLocaleDateString('en-IN', {
                      day: 'numeric',
                      month: 'short',
                      year: 'numeric',
                    })}
                  </p>
                </div>
                <span
                  className={`text-xs font-medium px-2.5 py-1 rounded-full capitalize ${
                    STATUS_COLORS[o.status] ?? 'bg-line text-ink/60'
                  }`}
                >
                  {o.status}
                </span>
              </div>

              <p className="text-sm text-ink/60 mb-3">
                {o.items?.length ?? 0} item{(o.items?.length ?? 0) !== 1 ? 's' : ''} ·{' '}
                <span className="font-mono">₹{o.total_amount.toFixed(2)}</span> ·{' '}
                {o.payment_method === 'cod' ? 'Cash on Delivery' : 'Paid online'}
              </p>

              <div className="flex items-center gap-4">
                <Link to={`/orders/${o.id}`} className="text-sm text-marigold font-medium hover:underline">
                  View details
                </Link>
                {(o.status === 'pending' || o.status === 'confirmed') && (
                  <button
                    onClick={() => handleCancel(o.id)}
                    disabled={cancellingId === o.id}
                    className="text-sm text-clay hover:underline disabled:opacity-50"
                  >
                    {cancellingId === o.id ? 'Cancelling...' : 'Cancel order'}
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
