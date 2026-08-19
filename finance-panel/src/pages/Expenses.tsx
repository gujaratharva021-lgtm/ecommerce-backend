import { useEffect, useState } from 'react'
import { listExpenses, createExpense, updateExpense, deleteExpense } from '../api/finance'
import { listWarehouses } from '../api/admin'
import type { Expense, ExpenseFormInput } from '../types/finance'
import { EXPENSE_CATEGORIES } from '../types/finance'
import type { Warehouse } from '../types/admin'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

const emptyForm: ExpenseFormInput = {
  amount: 0,
  category: 'misc',
  expense_date: todayISO(),
  warehouse_id: undefined,
  note: '',
  receipt_url: '',
}

export default function Expenses() {
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [warehouses, setWarehouses] = useState<Warehouse[]>([])
  const [totalAmount, setTotalAmount] = useState(0)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [page, setPage] = useState(1)
  const [categoryFilter, setCategoryFilter] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<ExpenseFormInput>(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listExpenses({
        category: categoryFilter || undefined,
        page,
        limit: 20,
      })
      setExpenses(res.expenses ?? [])
      setTotalAmount(res.total_amount)
      setTotal(res.total)
      setTotalPages(res.total_pages || 1)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load expenses.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [categoryFilter, page])

  useEffect(() => {
    listWarehouses()
      .then((res) => setWarehouses(res.warehouses ?? res ?? []))
      .catch(() => {})
  }, [])

  function openCreate() {
    setEditingId(null)
    setForm(emptyForm)
    setFormError(null)
    setShowForm(true)
  }

  function openEdit(exp: Expense) {
    setEditingId(exp.id)
    setForm({
      amount: exp.amount,
      category: exp.category,
      expense_date: exp.expense_date.slice(0, 10),
      warehouse_id: exp.warehouse_id ?? undefined,
      note: exp.note ?? '',
      receipt_url: exp.receipt_url ?? '',
    })
    setFormError(null)
    setShowForm(true)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setIsSaving(true)
    setFormError(null)
    try {
      if (editingId) {
        await updateExpense(editingId, form)
      } else {
        await createExpense(form)
      }
      setShowForm(false)
      await load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to save expense.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Delete this expense?')) return
    try {
      await deleteExpense(id)
      await load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to delete expense.')
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Expenses</h1>
          <p className="text-sm text-slate-500">
            {total} expense{total !== 1 ? 's' : ''} &middot; Total: {formatCurrency(totalAmount)}
          </p>
        </div>
        <button
          onClick={openCreate}
          className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium"
        >
          Add Expense
        </button>
      </div>

      <div className="flex items-center gap-2 mb-4 text-sm">
        <select
          value={categoryFilter}
          onChange={(e) => { setPage(1); setCategoryFilter(e.target.value) }}
          className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5"
        >
          <option value="">All categories</option>
          {EXPENSE_CATEGORIES.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading expenses...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500 mb-4">{error}</div>
      )}

      {!isLoading && !error && (
        <>
          <div className="border border-slate-800 rounded-xl overflow-hidden mb-4">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Date</th>
                  <th className="px-4 py-2 font-medium">Category</th>
                  <th className="px-4 py-2 font-medium">Warehouse</th>
                  <th className="px-4 py-2 font-medium">Note</th>
                  <th className="px-4 py-2 font-medium text-right">Amount</th>
                  <th className="px-4 py-2 font-medium text-right"></th>
                </tr>
              </thead>
              <tbody>
                {expenses.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-4 py-6 text-center text-slate-500">
                      No expenses recorded yet.
                    </td>
                  </tr>
                )}
                {expenses.map((exp) => (
                  <tr key={exp.id} className="border-t border-slate-800">
                    <td className="px-4 py-2">{new Date(exp.expense_date).toLocaleDateString('en-IN')}</td>
                    <td className="px-4 py-2 capitalize">{exp.category}</td>
                    <td className="px-4 py-2 text-slate-400">{exp.warehouse?.name ?? '-'}</td>
                    <td className="px-4 py-2 text-slate-400">{exp.note || '-'}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(exp.amount)}</td>
                    <td className="px-4 py-2 text-right space-x-3">
                      <button onClick={() => openEdit(exp)} className="text-emerald-400 hover:text-emerald-300 text-xs">
                        Edit
                      </button>
                      <button onClick={() => handleDelete(exp.id)} className="text-red-400 hover:text-red-300 text-xs">
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
              {editingId ? 'Edit Expense' : 'Add Expense'}
            </h2>

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

            <div>
              <label className="text-xs text-slate-400 block mb-1">Category</label>
              <select
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                {EXPENSE_CATEGORIES.map((c) => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Date</label>
              <input
                type="date"
                required
                value={form.expense_date}
                onChange={(e) => setForm({ ...form, expense_date: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Warehouse (optional)</label>
              <select
                value={form.warehouse_id ?? ''}
                onChange={(e) => setForm({ ...form, warehouse_id: e.target.value ? Number(e.target.value) : undefined })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                <option value="">Not warehouse-specific</option>
                {warehouses.map((w) => (
                  <option key={w.id} value={w.id}>{w.name}</option>
                ))}
              </select>
            </div>

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
