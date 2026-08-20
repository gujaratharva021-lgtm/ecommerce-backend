import { useEffect, useState } from 'react'
import {
  getBankTransactions,
  createBankTransaction,
  matchBankTransaction,
  ignoreBankTransaction,
  voidBankTransaction,
} from '../api/finance'
import type { BankTransaction, BankTransactionRequest } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(value)
}

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

const emptyForm: BankTransactionRequest = { transaction_date: todayISO(), description: '', amount: 0, reference_number: '' }

export default function BankReconciliation() {
  const [transactions, setTransactions] = useState<BankTransaction[]>([])
  const [unmatchedCount, setUnmatchedCount] = useState(0)
  const [statusFilter, setStatusFilter] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<BankTransactionRequest>(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [matchingId, setMatchingId] = useState<number | null>(null)
  const [matchType, setMatchType] = useState('')
  const [matchNote, setMatchNote] = useState('')

  function load() {
    setIsLoading(true)
    setError(null)
    getBankTransactions({ status: statusFilter || undefined })
      .then((res) => {
        setTransactions(res.transactions ?? [])
        setUnmatchedCount(res.unmatched_count ?? 0)
      })
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load bank transactions.'))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [statusFilter])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!form.transaction_date || !form.amount) {
      setFormError('Transaction date and amount are required.')
      return
    }
    setIsSaving(true)
    setFormError(null)
    try {
      await createBankTransaction(form)
      setForm(emptyForm)
      setShowForm(false)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Could not create bank transaction.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleMatch(id: number) {
    if (!matchType.trim()) return
    try {
      await matchBankTransaction(id, { matched_type: matchType, note: matchNote })
      setMatchingId(null)
      setMatchType('')
      setMatchNote('')
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not match transaction.')
    }
  }

  async function handleIgnore(id: number) {
    try {
      await ignoreBankTransaction(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not ignore transaction.')
    }
  }

  async function handleVoid(id: number) {
    const reason = prompt('Void this bank transaction? Enter a reason:')
    if (reason === null) return
    if (!reason.trim()) {
      alert('A reason is required to void a transaction.')
      return
    }
    try {
      await voidBankTransaction(id, { reason })
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not void transaction.')
    }
  }


  function statusBadge(status: string) {
    const cls =
      status === 'matched'
        ? 'bg-emerald-600/15 text-emerald-400'
        : status === 'ignored'
        ? 'bg-slate-800 text-slate-500'
        : status === 'voided'
        ? 'bg-slate-700/40 text-slate-500'
        : 'bg-amber-600/15 text-amber-400'
    return <span className={`text-xs px-2 py-0.5 rounded-full capitalize ${cls}`}>{status}</span>
  }


  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Bank Reconciliation</h1>
          <p className="text-sm text-slate-500">Match bank statement lines against internal records.</p>
        </div>
        <button
          onClick={() => setShowForm((s) => !s)}
          className="text-sm bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg transition-colors"
        >
          {showForm ? 'Cancel' : '+ New Transaction'}
        </button>
      </div>

      <div className="border border-amber-900 bg-amber-950/20 rounded-lg px-4 py-2 mb-6 text-xs text-amber-300 inline-block">
        {unmatchedCount} unmatched transaction{unmatchedCount === 1 ? '' : 's'} need review
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="border border-slate-800 rounded-xl p-5 mb-6 max-w-2xl">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <Field label="Transaction Date *">
              <input
                type="date"
                value={form.transaction_date}
                onChange={(e) => setForm({ ...form, transaction_date: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Amount * (negative = money out)">
              <input
                type="number"
                value={form.amount || ''}
                onChange={(e) => setForm({ ...form, amount: Number(e.target.value) })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Description">
              <input
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Reference Number">
              <input
                value={form.reference_number}
                onChange={(e) => setForm({ ...form, reference_number: e.target.value })}
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
            {isSaving ? 'Saving...' : 'Add Transaction'}
          </button>
        </form>
      )}

      <div className="flex gap-2 mb-4 text-sm">
        {['', 'unmatched', 'matched', 'ignored'].map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-3 py-1.5 rounded-lg capitalize transition-colors ${
              statusFilter === s ? 'bg-emerald-600/15 text-emerald-400' : 'text-slate-400 hover:bg-slate-800'
            }`}
          >
            {s === '' ? 'All' : s}
          </button>
        ))}
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading transactions...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && (
        <div className="border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-900 text-slate-400 text-left">
                <th className="px-4 py-2 font-medium">Date</th>
                <th className="px-4 py-2 font-medium">Description</th>
                <th className="px-4 py-2 font-medium">Ref #</th>
                <th className="px-4 py-2 font-medium text-right">Amount</th>
                <th className="px-4 py-2 font-medium">Status</th>
                <th className="px-4 py-2 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {transactions.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-6 text-center text-slate-500">
                    No transactions found.
                  </td>
                </tr>
              )}
              {transactions.map((t) => (
                <tr key={t.id} className="border-t border-slate-800">
                  <td className="px-4 py-2 text-slate-400">{t.transaction_date.slice(0, 10)}</td>
                  <td className="px-4 py-2">{t.description || 'ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â'}</td>
                  <td className="px-4 py-2 text-slate-400">{t.reference_number || 'ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â'}</td>
                  <td className={`px-4 py-2 text-right ${t.amount < 0 ? 'text-red-400' : ''}`}>
                    {formatCurrency(t.amount)}
                  </td>
                  <td className="px-4 py-2">
                    {statusBadge(t.status)}
                    {t.status === 'matched' && t.matched_type && (
                      <p className="text-xs text-slate-500 mt-0.5">{t.matched_type}</p>
                    )}
                  </td>
                  <td className="px-4 py-2 text-right space-x-2">
                    {t.status === 'unmatched' &&
                      (matchingId === t.id ? (
                        <span className="inline-flex items-center gap-1">
                          <input
                            value={matchType}
                            onChange={(e) => setMatchType(e.target.value)}
                            placeholder="Matched type"
                            className="w-32 bg-slate-800 border border-slate-700 rounded px-2 py-1 text-xs"
                          />
                          <input
                            value={matchNote}
                            onChange={(e) => setMatchNote(e.target.value)}
                            placeholder="Note"
                            className="w-24 bg-slate-800 border border-slate-700 rounded px-2 py-1 text-xs"
                          />
                          <button
                            onClick={() => handleMatch(t.id)}
                            className="text-xs text-emerald-400 hover:text-emerald-300"
                          >
                            Confirm
                          </button>
                          <button
                            onClick={() => {
                              setMatchingId(null)
                              setMatchType('')
                              setMatchNote('')
                            }}
                            className="text-xs text-slate-500 hover:text-slate-300"
                          >
                            Cancel
                          </button>
                        </span>
                      ) : (
                        <>
                          <button
                            onClick={() => setMatchingId(t.id)}
                            className="text-xs text-emerald-400 hover:text-emerald-300"
                          >
                            Match
                          </button>
                          <button
                            onClick={() => handleIgnore(t.id)}
                            className="text-xs text-slate-500 hover:text-slate-300"
                          >
                            Ignore
                          </button>
                        </>
                      ))}
                    <button
                      onClick={() => handleVoid(t.id)}
                      className="text-xs text-slate-500 hover:text-red-400 transition-colors"
                    >
                      Void
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
