import { useEffect, useState } from 'react'
import {
  listAdminPayments,
  getAdminPaymentDetail,
  settleGatewayPayment,
  listRiderPayouts,
  createRiderPayout,
  approveRiderPayout,
  payRiderPayout,
  listRiderCODDeposits,
  createRiderCODDeposit,
  verifyRiderCODDeposit,
} from '../api/finance'
import type { AdminPaymentRow, RiderPayout, RiderCODDeposit } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(value)
}

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

function daysAgoISO(days: number) {
  const d = new Date()
  d.setDate(d.getDate() - days)
  return d.toISOString().slice(0, 10)
}

type TabKey = 'gateway' | 'payouts' | 'deposits'

export default function Operations() {
  const [tab, setTab] = useState<TabKey>('gateway')

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-lg font-semibold">Operations</h1>
        <p className="text-sm text-slate-500">Gateway settlement, rider payouts, and rider COD deposits.</p>
      </div>

      <div className="flex gap-1 mb-6 border-b border-slate-800">
        {([
          { key: 'gateway', label: 'Gateway Settlement' },
          { key: 'payouts', label: 'Rider Payouts' },
          { key: 'deposits', label: 'Rider COD Deposits' },
        ] as { key: TabKey; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-3 py-2 text-sm rounded-t-lg transition-colors ${
              tab === t.key ? 'bg-slate-900 text-emerald-400 border-b-2 border-emerald-500' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'gateway' && <GatewaySettlementTab />}
      {tab === 'payouts' && <RiderPayoutsTab />}
      {tab === 'deposits' && <RiderCODDepositsTab />}
    </div>
  )
}

function GatewaySettlementTab() {
  const [payments, setPayments] = useState<AdminPaymentRow[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [settlingOrderId, setSettlingOrderId] = useState<number | null>(null)

  function load() {
    setIsLoading(true)
    setError(null)
    listAdminPayments({ status: 'paid', gateway: 'razorpay', limit: 50 })
      .then((res) => setPayments(res.payments))
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load payments.'))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function handleSettle(orderId: number) {
    setSettlingOrderId(orderId)
    try {
      const detail = await getAdminPaymentDetail(orderId)
      if (detail.payment.is_settled) {
        alert('This payment is already settled.')
        return
      }
      await settleGatewayPayment(detail.payment.id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not settle payment.')
    } finally {
      setSettlingOrderId(null)
    }
  }

  if (isLoading) return <p className="text-sm text-slate-500">Loading payments...</p>
  if (error) return <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>

  return (
    <>
      <div className="border border-slate-800 bg-slate-900/40 text-slate-400 text-xs rounded-lg px-4 py-3 mb-4">
        Gateway fee is an assumed flat 2% rate (not fetched from Razorpay) - see backend for details.
      </div>
      {payments.length === 0 ? (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">No paid online payments found.</div>
      ) : (
        <div className="border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-900 text-slate-400 text-xs uppercase">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Order</th>
                <th className="px-4 py-2 text-left font-medium">Customer</th>
                <th className="px-4 py-2 text-right font-medium">Amount</th>
                <th className="px-4 py-2 text-right font-medium">Action</th>
              </tr>
            </thead>
            <tbody>
              {payments.map((p) => (
                <tr key={p.order_id} className="border-t border-slate-800">
                  <td className="px-4 py-2">#{p.order_id}</td>
                  <td className="px-4 py-2">{p.customer_name || p.customer_phone}</td>
                  <td className="px-4 py-2 text-right">{formatCurrency(p.amount)}</td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={() => handleSettle(p.order_id)}
                      disabled={settlingOrderId === p.order_id}
                      className="text-xs text-emerald-400 hover:text-emerald-300 disabled:opacity-50"
                    >
                      {settlingOrderId === p.order_id ? 'Settling...' : 'Settle'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}

function RiderPayoutsTab() {
  const [payouts, setPayouts] = useState<RiderPayout[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [partnerId, setPartnerId] = useState('')
  const [periodFrom, setPeriodFrom] = useState(daysAgoISO(6))
  const [periodTo, setPeriodTo] = useState(todayISO())
  const [isCreating, setIsCreating] = useState(false)

  function load() {
    setIsLoading(true)
    listRiderPayouts({})
      .then((res) => setPayouts(res.rider_payouts))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCreate() {
    if (!partnerId) {
      alert('Enter a delivery partner ID.')
      return
    }
    setIsCreating(true)
    try {
      await createRiderPayout({ delivery_partner_id: Number(partnerId), period_from: periodFrom, period_to: periodTo })
      setPartnerId('')
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not create payout.')
    } finally {
      setIsCreating(false)
    }
  }

  async function handleApprove(id: number) {
    try {
      await approveRiderPayout(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not approve payout.')
    }
  }

  async function handlePay(id: number) {
    try {
      await payRiderPayout(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not pay payout.')
    }
  }

  function statusBadge(status: string) {
    const cls =
      status === 'paid' ? 'bg-emerald-600/15 text-emerald-400' : status === 'approved' ? 'bg-amber-600/15 text-amber-400' : 'bg-slate-800 text-slate-400'
    return <span className={`text-xs px-2 py-0.5 rounded-full ${cls}`}>{status}</span>
  }

  return (
    <>
      <div className="border border-slate-800 rounded-xl p-4 mb-6 flex items-end gap-3 flex-wrap">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Delivery Partner ID</label>
          <input value={partnerId} onChange={(e) => setPartnerId(e.target.value)} type="number" className="w-40 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm" />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">From</label>
          <input value={periodFrom} onChange={(e) => setPeriodFrom(e.target.value)} type="date" max={periodTo} className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-sm" />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">To</label>
          <input value={periodTo} onChange={(e) => setPeriodTo(e.target.value)} type="date" min={periodFrom} max={todayISO()} className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-sm" />
        </div>
        <button onClick={handleCreate} disabled={isCreating} className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50">
          {isCreating ? 'Computing...' : 'Create Payout'}
        </button>
      </div>

      {isLoading ? (
        <p className="text-sm text-slate-500">Loading payouts...</p>
      ) : payouts.length === 0 ? (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">No payouts yet.</div>
      ) : (
        <div className="border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-900 text-slate-400 text-xs uppercase">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Partner ID</th>
                <th className="px-4 py-2 text-left font-medium">Period</th>
                <th className="px-4 py-2 text-right font-medium">Deliveries</th>
                <th className="px-4 py-2 text-right font-medium">Amount</th>
                <th className="px-4 py-2 text-left font-medium">Status</th>
                <th className="px-4 py-2 text-right font-medium">Action</th>
              </tr>
            </thead>
            <tbody>
              {payouts.map((p) => (
                <tr key={p.id} className="border-t border-slate-800">
                  <td className="px-4 py-2">{p.delivery_partner_id}</td>
                  <td className="px-4 py-2 text-xs text-slate-400">{p.period_from.slice(0, 10)} to {p.period_to.slice(0, 10)}</td>
                  <td className="px-4 py-2 text-right">{p.delivered_count}</td>
                  <td className="px-4 py-2 text-right">{formatCurrency(p.amount)}</td>
                  <td className="px-4 py-2">{statusBadge(p.status)}</td>
                  <td className="px-4 py-2 text-right space-x-2">
                    {p.status === 'pending' && (
                      <button onClick={() => handleApprove(p.id)} className="text-xs text-amber-400 hover:text-amber-300">Approve</button>
                    )}
                    {p.status === 'approved' && (
                      <button onClick={() => handlePay(p.id)} className="text-xs text-emerald-400 hover:text-emerald-300">Pay</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}

function RiderCODDepositsTab() {
  const [deposits, setDeposits] = useState<RiderCODDeposit[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [partnerId, setPartnerId] = useState('')
  const [amount, setAmount] = useState('')
  const [depositDate, setDepositDate] = useState(todayISO())
  const [isCreating, setIsCreating] = useState(false)

  function load() {
    setIsLoading(true)
    listRiderCODDeposits({})
      .then((res) => setDeposits(res.rider_cod_deposits))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCreate() {
    if (!partnerId || !amount) {
      alert('Enter partner ID and amount.')
      return
    }
    setIsCreating(true)
    try {
      await createRiderCODDeposit({ delivery_partner_id: Number(partnerId), amount: Number(amount), deposit_date: depositDate })
      setPartnerId('')
      setAmount('')
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not record deposit.')
    } finally {
      setIsCreating(false)
    }
  }

  async function handleVerify(id: number) {
    try {
      await verifyRiderCODDeposit(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not verify deposit.')
    }
  }

  return (
    <>
      <div className="border border-slate-800 rounded-xl p-4 mb-6 flex items-end gap-3 flex-wrap">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Delivery Partner ID</label>
          <input value={partnerId} onChange={(e) => setPartnerId(e.target.value)} type="number" className="w-40 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm" />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Amount</label>
          <input value={amount} onChange={(e) => setAmount(e.target.value)} type="number" className="w-32 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm" />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Date</label>
          <input value={depositDate} onChange={(e) => setDepositDate(e.target.value)} type="date" max={todayISO()} className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-sm" />
        </div>
        <button onClick={handleCreate} disabled={isCreating} className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50">
          {isCreating ? 'Recording...' : 'Record Deposit'}
        </button>
      </div>

      {isLoading ? (
        <p className="text-sm text-slate-500">Loading deposits...</p>
      ) : deposits.length === 0 ? (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">No deposits yet.</div>
      ) : (
        <div className="border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-900 text-slate-400 text-xs uppercase">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Partner ID</th>
                <th className="px-4 py-2 text-left font-medium">Date</th>
                <th className="px-4 py-2 text-right font-medium">Amount</th>
                <th className="px-4 py-2 text-left font-medium">Status</th>
                <th className="px-4 py-2 text-right font-medium">Action</th>
              </tr>
            </thead>
            <tbody>
              {deposits.map((d) => (
                <tr key={d.id} className="border-t border-slate-800">
                  <td className="px-4 py-2">{d.delivery_partner_id}</td>
                  <td className="px-4 py-2 text-xs text-slate-400">{d.deposit_date.slice(0, 10)}</td>
                  <td className="px-4 py-2 text-right">{formatCurrency(d.amount)}</td>
                  <td className="px-4 py-2">
                    <span className={`text-xs px-2 py-0.5 rounded-full ${d.status === 'verified' ? 'bg-emerald-600/15 text-emerald-400' : 'bg-slate-800 text-slate-400'}`}>
                      {d.status}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-right">
                    {d.status === 'pending' && (
                      <button onClick={() => handleVerify(d.id)} className="text-xs text-emerald-400 hover:text-emerald-300">Verify</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}
