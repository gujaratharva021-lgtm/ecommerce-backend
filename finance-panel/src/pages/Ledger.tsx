import { useEffect, useState } from 'react'
import { getLedgerEntries, createManualJournalEntry, getTrialBalance, getAccounts } from '../api/finance'
import type { LedgerEntry, LedgerEntryLine, TrialBalance, Account } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(value)
}

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

const emptyLine = (): LedgerEntryLine => ({ account_id: 0, type: 'debit', amount: 0, description: '' })

export default function Ledger() {
  const [tab, setTab] = useState<'entries' | 'trial-balance'>('entries')
  const [entries, setEntries] = useState<LedgerEntry[]>([])
  const [accounts, setAccounts] = useState<Account[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [entryDate, setEntryDate] = useState(todayISO())
  const [lines, setLines] = useState<LedgerEntryLine[]>([emptyLine(), emptyLine()])
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [trialBalance, setTrialBalance] = useState<TrialBalance | null>(null)
  const [tbLoading, setTbLoading] = useState(false)
  const [tbAsOf, setTbAsOf] = useState(todayISO())

  function loadEntries() {
    setIsLoading(true)
    setError(null)
    getLedgerEntries({})
      .then((res) => setEntries(res.entries ?? []))
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load ledger entries.'))
      .finally(() => setIsLoading(false))
  }

  function loadTrialBalance() {
    setTbLoading(true)
    getTrialBalance(tbAsOf)
      .then(setTrialBalance)
      .catch(() => setTrialBalance(null))
      .finally(() => setTbLoading(false))
  }

  useEffect(() => {
    loadEntries()
    getAccounts().then((res) => setAccounts(res.accounts ?? [])).catch(() => {})
  }, [])

  useEffect(() => {
    if (tab === 'trial-balance') loadTrialBalance()
  }, [tab, tbAsOf])

  const totalDebit = lines.reduce((sum, l) => (l.type === 'debit' ? sum + (l.amount || 0) : sum), 0)
  const totalCredit = lines.reduce((sum, l) => (l.type === 'credit' ? sum + (l.amount || 0) : sum), 0)
  const isBalanced = totalDebit === totalCredit && totalDebit > 0

  function updateLine(idx: number, patch: Partial<LedgerEntryLine>) {
    setLines((prev) => prev.map((l, i) => (i === idx ? { ...l, ...patch } : l)))
  }

  function addLine() {
    setLines((prev) => [...prev, emptyLine()])
  }

  function removeLine(idx: number) {
    setLines((prev) => prev.filter((_, i) => i !== idx))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (lines.some((l) => !l.account_id || !l.amount)) {
      setFormError('Every line needs an account and a non-zero amount.')
      return
    }
    if (!isBalanced) {
      setFormError('Total debits must equal total credits before submitting.')
      return
    }
    setIsSaving(true)
    setFormError(null)
    try {
      await createManualJournalEntry({ entry_date: entryDate, lines })
      setLines([emptyLine(), emptyLine()])
      setEntryDate(todayISO())
      setShowForm(false)
      loadEntries()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Could not create journal entry.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Ledger</h1>
          <p className="text-sm text-slate-500">Double-entry journal — every transaction balances debits and credits.</p>
        </div>
        {tab === 'entries' && (
          <button
            onClick={() => setShowForm((s) => !s)}
            className="text-sm bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg transition-colors"
          >
            {showForm ? 'Cancel' : '+ New Journal Entry'}
          </button>
        )}
      </div>

      <div className="flex gap-2 mb-6 text-sm border-b border-slate-800">
        <button
          onClick={() => setTab('entries')}
          className={`px-4 py-2 border-b-2 transition-colors ${
            tab === 'entries' ? 'border-emerald-500 text-emerald-400' : 'border-transparent text-slate-400'
          }`}
        >
          Entries
        </button>
        <button
          onClick={() => setTab('trial-balance')}
          className={`px-4 py-2 border-b-2 transition-colors ${
            tab === 'trial-balance' ? 'border-emerald-500 text-emerald-400' : 'border-transparent text-slate-400'
          }`}
        >
          Trial Balance
        </button>
      </div>

      {tab === 'entries' && (
        <>
          {showForm && (
            <form onSubmit={handleSubmit} className="border border-slate-800 rounded-xl p-5 mb-6 max-w-3xl">
              <div className="mb-4">
                <label className="block">
                  <span className="text-xs text-slate-500 mb-1 block">Entry Date</span>
                  <input
                    type="date"
                    value={entryDate}
                    onChange={(e) => setEntryDate(e.target.value)}
                    className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                  />
                </label>
              </div>

              <div className="space-y-3 mb-4">
                {lines.map((line, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    <select
                      value={line.account_id}
                      onChange={(e) => updateLine(idx, { account_id: Number(e.target.value) })}
                      className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                    >
                      <option value={0}>Select account...</option>
                      {accounts.map((a) => (
                        <option key={a.id} value={a.id}>
                          {a.code} — {a.name}
                        </option>
                      ))}
                    </select>
                    <select
                      value={line.type}
                      onChange={(e) => updateLine(idx, { type: e.target.value as 'debit' | 'credit' })}
                      className="w-28 bg-slate-800 border border-slate-700 rounded-lg px-2 py-2 text-sm"
                    >
                      <option value="debit">Debit</option>
                      <option value="credit">Credit</option>
                    </select>
                    <input
                      type="number"
                      value={line.amount || ''}
                      onChange={(e) => updateLine(idx, { amount: Number(e.target.value) })}
                      placeholder="Amount"
                      className="w-32 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                    />
                    <input
                      value={line.description}
                      onChange={(e) => updateLine(idx, { description: e.target.value })}
                      placeholder="Description"
                      className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                    />
                    {lines.length > 2 && (
                      <button
                        type="button"
                        onClick={() => removeLine(idx)}
                        className="text-slate-500 hover:text-red-400 text-sm px-2"
                      >
                        ✕
                      </button>
                    )}
                  </div>
                ))}
              </div>

              <button
                type="button"
                onClick={addLine}
                className="text-xs text-emerald-400 hover:text-emerald-300 mb-4"
              >
                + Add line
              </button>

              <div className="flex items-center justify-between border-t border-slate-800 pt-4 mb-4 text-sm">
                <span className="text-slate-400">
                  Debit: <span className="text-slate-100 font-medium">{formatCurrency(totalDebit)}</span> · Credit:{' '}
                  <span className="text-slate-100 font-medium">{formatCurrency(totalCredit)}</span>
                </span>
                <span className={isBalanced ? 'text-emerald-400' : 'text-amber-400'}>
                  {isBalanced ? '✓ Balanced' : 'Not balanced yet'}
                </span>
              </div>

              {formError && <p className="text-sm text-red-400 mb-3">{formError}</p>}
              <button
                type="submit"
                disabled={isSaving || !isBalanced}
                className="text-sm bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white px-4 py-2 rounded-lg transition-colors"
              >
                {isSaving ? 'Saving...' : 'Post Journal Entry'}
              </button>
            </form>
          )}

          {isLoading && <p className="text-sm text-slate-500">Loading entries...</p>}
          {!isLoading && error && (
            <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
          )}

          {!isLoading && !error && (
            <div className="border border-slate-800 rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-slate-900 text-slate-400 text-left">
                    <th className="px-4 py-2 font-medium">Date</th>
                    <th className="px-4 py-2 font-medium">Ref</th>
                    <th className="px-4 py-2 font-medium">Account</th>
                    <th className="px-4 py-2 font-medium">Description</th>
                    <th className="px-4 py-2 font-medium">Type</th>
                    <th className="px-4 py-2 font-medium text-right">Amount</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.length === 0 && (
                    <tr>
                      <td colSpan={6} className="px-4 py-6 text-center text-slate-500">
                        No ledger entries yet.
                      </td>
                    </tr>
                  )}
                  {entries.map((e) => (
                    <tr key={e.id} className="border-t border-slate-800">
                      <td className="px-4 py-2 text-slate-400">{e.entry_date.slice(0, 10)}</td>
                      <td className="px-4 py-2 font-mono text-xs text-slate-500">{e.transaction_ref}</td>
                      <td className="px-4 py-2">
                        {e.account ? `${e.account.code} — ${e.account.name}` : `#${e.account_id}`}
                      </td>
                      <td className="px-4 py-2 text-slate-400">{e.description || '—'}</td>
                      <td className="px-4 py-2 capitalize">{e.type}</td>
                      <td className={`px-4 py-2 text-right ${e.type === 'debit' ? '' : 'text-slate-400'}`}>
                        {formatCurrency(e.amount)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {tab === 'trial-balance' && (
        <>
          <div className="mb-4">
            <label className="text-sm">
              <span className="text-slate-500 mr-2">As of</span>
              <input
                type="date"
                value={tbAsOf}
                onChange={(e) => setTbAsOf(e.target.value)}
                className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5"
              />
            </label>
          </div>

          {tbLoading && <p className="text-sm text-slate-500">Loading trial balance...</p>}

          {!tbLoading && trialBalance && (
            <>
              <div
                className={`border rounded-lg px-4 py-3 mb-4 text-sm ${
                  trialBalance.is_balanced
                    ? 'border-emerald-900 bg-emerald-950/30 text-emerald-300'
                    : 'border-red-900 bg-red-950/30 text-red-300'
                }`}
              >
                {trialBalance.is_balanced
                  ? `✓ Balanced — total debits and credits both equal ${formatCurrency(trialBalance.total_debit)}.`
                  : `⚠ Not balanced — debits ${formatCurrency(trialBalance.total_debit)} vs credits ${formatCurrency(
                      trialBalance.total_credit
                    )}.`}
              </div>

              <div className="border border-slate-800 rounded-xl overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-slate-900 text-slate-400 text-left">
                      <th className="px-4 py-2 font-medium">Code</th>
                      <th className="px-4 py-2 font-medium">Account</th>
                      <th className="px-4 py-2 font-medium">Type</th>
                      <th className="px-4 py-2 font-medium text-right">Debit</th>
                      <th className="px-4 py-2 font-medium text-right">Credit</th>
                    </tr>
                  </thead>
                  <tbody>
                    {trialBalance.accounts.length === 0 && (
                      <tr>
                        <td colSpan={5} className="px-4 py-6 text-center text-slate-500">
                          No activity as of this date.
                        </td>
                      </tr>
                    )}
                    {trialBalance.accounts.map((row) => (
                      <tr key={row.account_id} className="border-t border-slate-800">
                        <td className="px-4 py-2 font-mono text-xs">{row.account_code}</td>
                        <td className="px-4 py-2 font-medium">{row.account_name}</td>
                        <td className="px-4 py-2 capitalize text-slate-400">{row.account_type}</td>
                        <td className="px-4 py-2 text-right">
                          {row.total_debit > 0 ? formatCurrency(row.total_debit) : '—'}
                        </td>
                        <td className="px-4 py-2 text-right">
                          {row.total_credit > 0 ? formatCurrency(row.total_credit) : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr className="border-t border-slate-700 bg-slate-900/50 font-semibold">
                      <td className="px-4 py-2" colSpan={3}>
                        Total
                      </td>
                      <td className="px-4 py-2 text-right">{formatCurrency(trialBalance.total_debit)}</td>
                      <td className="px-4 py-2 text-right">{formatCurrency(trialBalance.total_credit)}</td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}
