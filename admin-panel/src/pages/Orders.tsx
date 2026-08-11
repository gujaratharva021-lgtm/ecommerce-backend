import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { listOrders, updateOrderStatus, assignDeliveryPartner, listDeliveryPartners } from '../api/admin'
import type { Order, DeliveryPartner } from '../types/admin'

const STATUS_OPTIONS = [
  'pending',
  'confirmed',
  'shipped',
  'delivered',
  'cancelled',
]

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-amber-500/15 text-amber-300',
  confirmed: 'bg-blue-500/15 text-blue-300',
  shipped: 'bg-indigo-500/15 text-indigo-300',
  delivered: 'bg-emerald-500/15 text-emerald-300',
  cancelled: 'bg-red-500/15 text-red-300',
}

export default function Orders() {
  const [orders, setOrders] = useState<Order[]>([])
  const [partners, setPartners] = useState<DeliveryPartner[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('')
  const [assigningId, setAssigningId] = useState<number | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const [ordersRes, partnersRes] = await Promise.all([
        listOrders(statusFilter ? { status: statusFilter } : undefined),
        listDeliveryPartners(),
      ])
      setOrders(ordersRes.orders ?? [])
      setPartners(partnersRes.delivery_partners ?? partnersRes.partners ?? partnersRes ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load orders.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter])

  async function handleStatusChange(id: number, status: string) {
    try {
      await updateOrderStatus(id, status)
      setOrders((prev) => prev.map((o) => (o.id === id ? { ...o, status } : o)))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update order status.')
    }
  }

  async function handleAssignDelivery(orderId: number, partnerId: string) {
    if (!partnerId) return
    setAssigningId(orderId)
    try {
      await assignDeliveryPartner(orderId, parseInt(partnerId, 10))
      alert('Delivery partner assigned successfully.')
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to assign delivery partner.')
    } finally {
      setAssigningId(null)
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Orders</h1>
            <p className="text-sm text-slate-400 mt-1">
              {orders.length} order{orders.length !== 1 ? 's' : ''}
            </p>
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
          >
            <option value="">All statuses</option>
            {STATUS_OPTIONS.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && orders.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No orders {statusFilter ? `with status "${statusFilter}"` : 'yet'}.
          </div>
        )}

        {!isLoading && orders.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Order</th>
                  <th className="px-4 py-3 font-medium">User</th>
                  <th className="px-4 py-3 font-medium">Products</th>
                  <th className="px-4 py-3 font-medium">Total</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Update</th>
                  <th className="px-4 py-3 font-medium">Assign Delivery</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((o) => (
                  <tr key={o.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">#{o.id}</td>
                    <td className="px-4 py-3">User {o.user_id}</td>
                    <td className="px-4 py-3 max-w-xs">
                      {o.items && o.items.length > 0 ? (
                        <div className="space-y-0.5">
                          {o.items.map((it) => (
                            <div key={it.id} className="text-slate-300">
                              {it.product?.name ?? `Product #${it.product_id}`}
                              <span className="text-slate-500"> × {it.quantity}</span>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <span className="text-xs text-slate-500">No items</span>
                      )}
                    </td>
                    <td className="px-4 py-3">₹{o.total_amount}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-1 rounded-md text-xs font-medium ${
                          STATUS_COLORS[o.status] ?? 'bg-slate-700 text-slate-300'
                        }`}
                      >
                        {o.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <select
                        defaultValue=""
                        onChange={(e) => {
                          if (e.target.value) handleStatusChange(o.id, e.target.value)
                        }}
                        className="bg-slate-800 border border-slate-700 rounded-md px-2 py-1 text-xs"
                      >
                        <option value="">Change status...</option>
                        {STATUS_OPTIONS.map((s) => (
                          <option key={s} value={s}>
                            {s}
                          </option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-3">
                      {(o.status === 'confirmed' || o.status === 'shipped') ? (
                        <select
                          defaultValue=""
                          disabled={assigningId === o.id}
                          onChange={(e) => handleAssignDelivery(o.id, e.target.value)}
                          className="bg-slate-800 border border-slate-700 rounded-md px-2 py-1 text-xs disabled:opacity-50"
                        >
                          <option value="">
                            {assigningId === o.id ? 'Assigning...' : 'Assign partner...'}
                          </option>
                          {partners.map((p) => (
                            <option key={p.id} value={p.id}>
                              {p.name} ({p.phone})
                            </option>
                          ))}
                        </select>
                      ) : (
                        <span className="text-xs text-slate-500">Confirm order first</span>
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
