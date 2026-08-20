import { useEffect, useState } from 'react'
import { getVendorBills, createVendorBill, payVendorBill, deleteVendorBill, getVendors } from '../api/finance'
import type { VendorBill, VendorBillRequest, Vendor } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(value)
}

const emptyForm: VendorBillRequest = { vendor_id: 0, bill_number: '', amount: 0, gst_amount: 0, bill_date: '', due_date: '', note: '' }

export default function VendorBills() {
  const [bills, setBills] = useState<VendorBill[]>([])
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [totalOutstanding, setTotalOutstanding] = useState(0)
  const [statusFilter, setStatusFilter] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<VendorBillRequest>(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [payingId, setPayingId] = useState<number | null>(null)
  const [payAmount, setPayAmount] = useState('')

  function load() {
    setIsLoading(true)
    setError(null)
    getVendorBills({ status: statusFilter || undefined })
      .then((res) => {
        setBills(res.bills ?? [])
        setTotalOutstanding(res.total_outstanding ?? 0)
      })
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load vendor bills.'))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [statusFilter])

  useEffect(() => {
    getVendors().then((res) => setVendors(res.vendors ?? [])).catch(() => {})
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!form.vendor_id || !form.amount || !form.bill_date) {
      setFormError('Vendor, amount, and bill date are required.')
      return
    }
    setIsSaving(true)
    setFormError(null)
    try {
      await createVendorBill(form)
      setForm(emptyForm)
      setShowForm(false)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Could not create vendor bill.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handlePay(id: number) {
    const amount = parseFloat(payAmount)
    if (!amount || amount <= 0) return
    try {
      await payVendorBill(id, amount)
      setPayingId(null)
      setPayAmount('')
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not record payment.')
    }
  }

  async function handleDelete(id: number, billNumber?: string) {
    if (!confirm(`Delete bill "${billNumber || id}"? This cannot be undone.`)) return
    try {
      await deleteVendorBill(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not delete bill.')
    }
  }

  function statusBadge(status: string) {
    const cls =
      status === 'paid'
        ? 'bg-emerald-600/15 text-emerald-400'
        : status === 'partially_paid'
        ? 'bg-amber-600/15 text-amber-400'
        : 'bg-red-600/15 text-red-400'
    const labelMap: Record<string, string> = { paid: 'Paid', partially_paid: 'Partially Paid', unpaid: 'Unpaid' }
    return <span className={`text-xs px-2 py-0.5 rounded-full ${cls}`}>{labelMap[status] ?? status}</span>
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Vendor Bills</h1>
          <p className="text-sm text-slate-500">Accounts payable Ã¢â‚¬â€ bills raised by vendors.</p>
        </div>
        <button
          onClick={() => setShowForm((s) => !s)}
          className="text-sm bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg transition-colors"
        >
          {showForm ? 'Cancel' : '+ New Bill'}
        </button>
      </div>

      <div className="border border-slate-800 rounded-xl p-4 mb-6 max-w-xs">
        <p className="text-xs text-slate-500 mb-1">Total Outstanding</p>
        <p className="text-lg font-semibold text-amber-400">{formatCurrency(totalOutstanding)}</p>
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="border border-slate-800 rounded-xl p-5 mb-6 max-w-2xl">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <Field label="Vendor *">
              <select
                value={form.vendor_id}
                onChange={(e) => setForm({ ...form, vendor_id: Number(e.target.value) })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                <option value={0}>Select vendor...</option>
                {vendors.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="Bill Number">
              <input
                value={form.bill_number}
                onChange={(e) => setForm({ ...form, bill_number: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Amount *">
              <input
                type="number"
                value={form.amount || ''}
                onChange={(e) => setForm({ ...form, amount: Number(e.target.value) })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="GST Amount">
              <input
                type="number"
                value={form.gst_amount || ''}
                onChange={(e) => setForm({ ...form, gst_amount: Number(e.target.value) })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Bill Date *">
              <input
                type="date"
                value={form.bill_date}
                onChange={(e) => setForm({ ...form, bill_date: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Due Date">
              <input
                type="date"
                value={form.due_date}
                onChange={(e) => setForm({ ...form, due_date: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Note">
              <input
                value={form.note}
                onChange={(e) => setForm({ ...form, note: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
          </div>
          {formError && <p className="text-sm text-red-400 mb-3">{formError}</p>}
          <button
            type="submit"
            disabled={isSaving}
            className="text-sm bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white px-4 py-2 rounded-lg transition-colors"
          >
            {isSaving ? 'Saving...' : 'Create Bill'}
          </button>
        </form>
      )}

      <div className="flex gap-2 mb-4 text-sm">
        {['', 'unpaid', 'partially_paid', 'paid'].map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-3 py-1.5 rounded-lg transition-colors ${
              statusFilter === s ? 'bg-emerald-600/15 text-emerald-400' : 'text-slate-400 hover:bg-slate-800'
            }`}
          >
            {s === '' ? 'All' : s === 'unpaid' ? 'Unpaid' : s === 'partially_paid' ? 'Partially Paid' : 'Paid'}
          </button>
        ))}
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading bills...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && (
        <div className="border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-900 text-slate-400 text-left">
                <th className="px-4 py-2 font-medium">Vendor</th>
                <th className="px-4 py-2 font-medium">Bill #</th>
                <th className="px-4 py-2 font-medium">Date</th>
                <th className="px-4 py-2 font-medium text-right">Amount</th>
                <th className="px-4 py-2 font-medium text-right">GST</th>
                <th className="px-4 py-2 font-medium text-right">Paid</th>
                <th className="px-4 py-2 font-medium">Status</th>
                <th className="px-4 py-2 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {bills.length === 0 && (
                <tr>
                  <td colSpan={8} className="px-4 py-6 text-center text-slate-500">
                    No bills found.
                  </td>
                </tr>
              )}
              {bills.map((b) => (
                <tr key={b.id} className="border-t border-slate-800">
                  <td className="px-4 py-2 font-medium">{b.vendor?.name ?? `#${b.vendor_id}`}</td>
                  <td className="px-4 py-2 text-slate-400">{b.bill_number || 'Ã¢â‚¬â€'}</td>
                  <td className="px-4 py-2 text-slate-400">{b.bill_date.slice(0, 10)}</td>
                  <td className="px-4 py-2 text-right">{formatCurrency(b.amount)}</td>
                  <td className="px-4 py-2 text-right text-slate-400">{formatCurrency(b.gst_amount)}</td>
                  <td className="px-4 py-2 text-right text-slate-400">{formatCurrency(b.amount_paid)}</td>
                  <td className="px-4 py-2">{statusBadge(b.status)}</td>
                  <td className="px-4 py-2 text-right space-x-2">
                    {b.status !== 'paid' &&
                      (payingId === b.id ? (
                        <span className="inline-flex items-center gap-1">
                          <input
                            type="number"
                            value={payAmount}
                            onChange={(e) => setPayAmount(e.target.value)}
                            placeholder="Amount"
                            className="w-24 bg-slate-800 border border-slate-700 rounded px-2 py-1 text-xs"
                          />
                          <button
                            onClick={() => handlePay(b.id)}
                            className="text-xs text-emerald-400 hover:text-emerald-300"
                          >
                            Confirm
                          </button>
                          <button
                            onClick={() => {
                              setPayingId(null)
                              setPayAmount('')
                            }}
                            className="text-xs text-slate-500 hover:text-slate-300"
                          >
                            Cancel
                          </button>
                        </span>
                      ) : (
                        <button
                          onClick={() => setPayingId(b.id)}
                          className="text-xs text-emerald-400 hover:text-emerald-300"
                        >
                          Pay
                        </button>
                      ))}
                    <button
                      onClick={() => handleDelete(b.id, b.bill_number)}
                      className="text-xs text-slate-500 hover:text-red-400 transition-colors"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs text-slate-500 mb-1 block">{label}</span>
      {children}
    </label>
  )
}
