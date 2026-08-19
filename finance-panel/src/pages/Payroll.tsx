import { useEffect, useState } from 'react'
import { listPayroll, createPayroll, updatePayroll, deletePayroll } from '../api/finance'
import { listWarehouseStaff } from '../api/admin'
import type { Payroll, PayrollFormInput } from '../types/finance'
import { PAYROLL_PAYMENT_METHODS } from '../types/finance'
import type { WarehouseStaff } from '../types/admin'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

function currentMonth() {
  return new Date().getMonth() + 1
}
function currentYear() {
  return new Date().getFullYear()
}

const emptyForm: PayrollFormInput = {
  staff_id: 0,
  amount: 0,
  month: currentMonth(),
  year: currentYear(),
  status: 'pending',
  payment_method: '',
  note: '',
}

export default function Payroll() {
  const [records, setRecords] = useState<Payroll[]>([])
  const [staff, setStaff] = useState<WarehouseStaff[]>([])
  const [totalPending, setTotalPending] = useState(0)
  const [totalPaid, setTotalPaid] = useState(0)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<PayrollFormInput>(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listPayroll({
        status: statusFilter || undefined,
        page,
        limit: 20,
      })
      setRecords(res.payroll ?? [])
      setTotalPending(res.total_pending)
      setTotalPaid(res.total_paid)
      setTotal(res.total)
      setTotalPages(res.total_pages || 1)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load payroll.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [statusFilter, page])

  useEffect(() => {
    listWarehouseStaff()
      .then((res) => setStaff(res.warehouse_staff ?? res ?? []))
      .catch(() => {})
  }, [])

  function openCreate() {
    setEditingId(null)
    setForm(emptyForm)
    setFormError(null)
    setShowForm(true)
  }

  function openEdit(p: Payroll) {
    setEditingId(p.id)
    setForm({
      staff_id: p.staff_id,
      amount: p.amount,
      month: p.month,
      year: p.year,
      status: p.status,
      payment_method: p.payment_method ?? '',
      note: p.note ?? '',
    })
    setFormError(null)
    setShowForm(true)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.staff_id) {
      setFormError('Select a staff member.')
      return
    }
    setIsSaving(true)
    setFormError(null)
    try {
      if (editingId) {
        await updatePayroll(editingId, form)
      } else {
        await createPayroll(form)
      }
      setShowForm(false)
      await load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to save payroll record.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleMarkPaid(p: Payroll) {
    try {
      await updatePayroll(p.id, {
        staff_id: p.staff_id,
        amount: p.amount,
        month: p.month,
        year: p.year,
        status: 'paid',
        payment_method: p.payment_method || 'bank',
        note: p.note,
      })
      await load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to mark as paid.')
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Delete this payroll record?')) return
    try {
      await deletePayroll(id)
      await load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to delete payroll record.')
    }
  }

  function statusColor(status: string) {
    return status === 'paid' ? 'bg-emerald-500/15 text-emerald-400' : 'bg-amber-500/15 text-amber-400'
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Payroll</h1>
          <p className="text-sm text-slate-500">
            {total} record{total !== 1 ? 's' : ''} &middot; Pending: {formatCurrency(totalPending)} &middot; Paid: {formatCurrency(totalPaid)}
          </p>
        </div>
        <button
          onClick={openCreate}
          className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium"
        >
          Add Payroll
        </button>
      </div>

      <div className="flex items-center gap-2 mb-4 text-sm">
        <select
          value={statusFilter}
          onChange={(e) => { setPage(1); setStatusFilter(e.target.value) }}
          className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5"
        >
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="paid">Paid</option>
        </select>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading payroll...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500 mb-4">{error}</div>
      )}

      {!isLoading && !error && (
        <>
          <div className="border border-slate-800 rounded-xl overflow-hidden mb-4">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Staff</th>
                  <th className="px-4 py-2 font-medium">Period</th>
                  <th className="px-4 py-2 font-medium text-right">Amount</th>
                  <th className="px-4 py-2 font-medium">Status</th>
                  <th className="px-4 py-2 font-medium text-right"></th>
                </tr>
              </thead>
              <tbody>
                {records.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-4 py-6 text-center text-slate-500">
                      No payroll records yet.
                    </td>
                  </tr>
                )}
                {records.map((p) => (
                  <tr key={p.id} className="border-t border-slate-800">
                    <td className="px-4 py-2">{p.staff?.name ?? `Staff #${p.staff_id}`}</td>
                    <td className="px-4 py-2 text-slate-400">{MONTHS[p.month - 1]} {p.year}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(p.amount)}</td>
                    <td className="px-4 py-2">
                      <span className={`text-xs px-2 py-1 rounded-full ${statusColor(p.status)}`}>
                        {p.status}
                      </span>
                    </td>
                    <td className="px-4 py-2 text-right space-x-3">
                      {p.status === 'pending' && (
                        <button onClick={() => handleMarkPaid(p)} className="text-emerald-400 hover:text-emerald-300 text-xs">
                          Mark Paid
                        </button>
                      )}
                      <button onClick={() => openEdit(p)} className="text-slate-400 hover:text-slate-200 text-xs">
                        Edit
                      </button>
                      <button onClick={() => handleDelete(p.id)} className="text-red-400 hover:text-red-300 text-xs">
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between text-sm text-slate-500">
            <span>{total} total</span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-2 py-1 border border-slate-700 rounded disabled:opacity-40"
              >
                Prev
              </button>
              <span>Page {page} of {totalPages}</span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="px-2 py-1 border border-slate-700 rounded disabled:opacity-40"
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}

      {showForm && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <form
            onSubmit={handleSubmit}
            className="bg-slate-900 border border-slate-800 rounded-xl p-6 w-full max-w-md space-y-3"
          >
            <h2 className="text-base font-semibold mb-2">
              {editingId ? 'Edit Payroll' : 'Add Payroll'}
            </h2>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Staff</label>
              <select
                required
                value={form.staff_id || ''}
                onChange={(e) => setForm({ ...form, staff_id: Number(e.target.value) })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                <option value="">Select staff member</option>
                {staff.map((s) => (
                  <option key={s.id} value={s.id}>{s.name} ({s.phone})</option>
                ))}
              </select>
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Amount</label>
              <input
                type="number"
                min="0"
                step="0.01"
                required
                value={form.amount || ''}
                onChange={(e) => setForm({ ...form, amount: parseFloat(e.target.value) || 0 })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>

            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Month</label>
                <select
                  value={form.month}
                  onChange={(e) => setForm({ ...form, month: Number(e.target.value) })}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                >
                  {MONTHS.map((m, i) => (
                    <option key={m} value={i + 1}>{m}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs text-slate-400 block mb-1">Year</label>
                <input
                  type="number"
                  required
                  value={form.year}
                  onChange={(e) => setForm({ ...form, year: parseInt(e.target.value, 10) || currentYear() })}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Status</label>
              <select
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                <option value="pending">Pending</option>
                <option value="paid">Paid</option>
              </select>
            </div>

            {form.status === 'paid' && (
              <div>
                <label className="text-xs text-slate-400 block mb-1">Payment Method</label>
                <select
                  value={form.payment_method}
                  onChange={(e) => setForm({ ...form, payment_method: e.target.value })}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                >
                  <option value="">Select method</option>
                  {PAYROLL_PAYMENT_METHODS.map((m) => (
                    <option key={m} value={m}>{m}</option>
                  ))}
                </select>
              </div>
            )}

            <div>
              <label className="text-xs text-slate-400 block mb-1">Note (optional)</label>
              <input
                type="text"
                value={form.note}
                onChange={(e) => setForm({ ...form, note: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>

            {formError && <p className="text-red-400 text-xs">{formError}</p>}

            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs font-medium"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSaving}
                className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium disabled:opacity-50"
              >
                {isSaving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}

