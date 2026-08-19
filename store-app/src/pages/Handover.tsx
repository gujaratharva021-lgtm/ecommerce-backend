import { useCallback, useEffect, useState } from 'react'
import { listWarehouseOrders, handoverOrder as handoverOrderApi } from '../api/warehouse'
import type { Order } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

export default function Handover() {
  const [orders, setOrders] = useState<Order[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [handoverTarget, setHandoverTarget] = useState<Order | null>(null)
  const [packageCount, setPackageCount] = useState(1)
  const [handoverSubmitting, setHandoverSubmitting] = useState(false)
  const [handoverError, setHandoverError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await listWarehouseOrders({ status: 'ready_for_dispatch', page: 1, limit: 50 })
      setOrders(data.orders ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load orders ready for handover.'))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  function openHandover(order: Order) {
    setHandoverTarget(order)
    setPackageCount(1)
    setHandoverError(null)
  }

  async function handleConfirmHandover() {
    if (!handoverTarget || !handoverTarget.delivery_partner) return
    setHandoverSubmitting(true)
    setHandoverError(null)
    try {
      await handoverOrderApi(handoverTarget.id, {
        package_count: packageCount,
        delivery_partner_id: handoverTarget.delivery_partner.id,
      })
      setHandoverTarget(null)
      await load()
    } catch (err) {
      setHandoverError(getErrorMessage(err, 'Failed to record handover.'))
    } finally {
      setHandoverSubmitting(false)
    }
  }

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="font-display text-2xl font-semibold">Handover</h1>
          <p className="text-xs text-slate-500 mt-1">Orders ready for dispatch to a delivery partner.</p>
        </div>
        <button
          onClick={load}
          className="text-xs px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          Refresh
        </button>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading && <p className="text-sm text-slate-400">Loading orders...</p>}

      {!isLoading && orders.length === 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-8 text-center text-sm text-slate-500">
          No orders are ready for handover right now.
        </div>
      )}

      {!isLoading && orders.length > 0 && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-800/50 text-slate-400 text-xs uppercase">
              <tr>
                <th className="text-left px-4 py-2.5">Order</th>
                <th className="text-left px-4 py-2.5">Customer</th>
                <th className="text-left px-4 py-2.5">Delivery Partner</th>
                <th className="text-left px-4 py-2.5">Payment</th>
                <th className="text-right px-4 py-2.5">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {orders.map((order) => (
                <tr key={order.id} className="hover:bg-slate-800/30">
                  <td className="px-4 py-3">#{order.id}</td>
                  <td className="px-4 py-3 text-slate-300">{order.customer_name ?? '-'}</td>
                  <td className="px-4 py-3">
                    {order.delivery_partner ? (
                      <span className="text-slate-300">{order.delivery_partner.name}</span>
                    ) : (
                      <span className="text-rose-400 text-xs">Not assigned</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-slate-400 uppercase text-xs">{order.payment_method}</td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openHandover(order)}
                      className="px-3 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-medium transition-colors"
                    >
                      Handover
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {handoverTarget && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 w-full max-w-md">
            <h2 className="text-base font-semibold mb-4">Handover Order #{handoverTarget.id}</h2>
            {handoverTarget.delivery_partner ? (
              <div className="border border-slate-800 rounded-lg p-3 mb-4 text-sm">
                <p className="text-slate-400 text-xs uppercase mb-1">Assigned Delivery Partner</p>
                <p className="font-medium">{handoverTarget.delivery_partner.name}</p>
                <p className="text-slate-400">{handoverTarget.delivery_partner.phone}</p>
                {handoverTarget.delivery_partner.vehicle_number && (
                  <p className="text-slate-400 text-xs">Vehicle: {handoverTarget.delivery_partner.vehicle_number}</p>
                )}
                <p className="text-xs text-amber-400 mt-2">Verify this partner is physically present before confirming.</p>
              </div>
            ) : (
              <p className="text-sm text-rose-400 mb-4">No delivery partner assigned to this order yet. Cannot hand over.</p>
            )}

            <label className="block text-xs text-slate-400 mb-1">Package Count</label>
            <input
              type="number"
              min={1}
              value={packageCount}
              onChange={(e) => setPackageCount(parseInt(e.target.value, 10) || 1)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm mb-4"
            />

            {handoverError && (
              <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-3 py-2 mb-4">
                {handoverError}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <button
                onClick={() => setHandoverTarget(null)}
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirmHandover}
                disabled={handoverSubmitting || !handoverTarget.delivery_partner}
                className="px-3 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-medium transition-colors disabled:opacity-50"
              >
                {handoverSubmitting ? 'Confirming...' : 'Confirm Handover'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
