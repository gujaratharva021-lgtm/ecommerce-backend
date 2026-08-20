import { useEffect, useState } from 'react'
import { getAccounts, createAccount, updateAccount } from '../api/finance'
import type { Account, AccountRequest } from '../types/finance'

const ACCOUNT_TYPES = ['asset', 'liability', 'equity', 'revenue', 'expense']
const emptyForm: AccountRequest = { code: '', name: '', type: 'asset', is_active: true }

export default function Accounts() {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<AccountRequest>(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  function load() {
    setIsLoading(true)
    setError(null)
    getAccounts()
      .then((res) => setAccounts(res.accounts ?? []))
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load accounts.'))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  function startCreate() {
    setEditingId(null)
    setForm(emptyForm)
    setShowForm(true)
    setFormError(null)
  }

  function startEdit(a: Account) {
    setEditingId(a.id)
    setForm({ code: a.code, name: a.name, type: a.type, is_active: a.is_active })
    setShowForm(true)
    setFormError(null)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.code.trim() || !form.name.trim()) {
      setFormError('Code and name are required.')
      return
    }
    setIsSaving(true)
    setFormError(null)
    try {
      if (editingId) {
        await updateAccount(editingId, form)
      } else {
        await createAccount(form)
      }
      setForm(emptyForm)
      setShowForm(false)
      setEditingId(null)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Could not save account.')
    } finally {
      setIsSaving(false)
    }
  }

  const grouped = ACCOUNT_TYPES.map((type) => ({
    type,
    items: accounts.filter((a) => a.type === type),
  })).filter((g) => g.items.length > 0)

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Chart of Accounts</h1>
          <p className="text-sm text-slate-500">The fixed list of buckets every ledger entry debits or credits.</p>
        </div>
        <button
          onClick={() => (showForm ? setShowForm(false) : startCreate())}
          className="text-sm bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg transition-colors"
        >
          {showForm ? 'Cancel' : '+ New Account'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="border border-slate-800 rounded-xl p-5 mb-6 max-w-xl">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <Field label="Code *">
              <input
                value={form.code}
                onChange={(e) => setForm({ ...form, code: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Name *">
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Type *">
              <select
                value={form.type}
                onChange={(e) => setForm({ ...form, type: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                {ACCOUNT_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {t.charAt(0).toUpperCase() + t.slice(1)}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="Active">
              <select
                value={form.is_active ? 'true' : 'false'}
                onChange={(e) => setForm({ ...form, is_active: e.target.value === 'true' })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </Field>
          </div>
          {formError && <p className="text-sm text-red-400 mb-3">{formError}</p>}
          <button
            type="submit"
            disabled={isSaving}
            className="text-sm bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white px-4 py-2 rounded-lg transition-colors"
          >
            {isSaving ? 'Saving...' : editingId ? 'Update Account' : 'Create Account'}
          </button>
        </form>
      )}

      {isLoading && <p className="text-sm text-slate-500">Loading accounts...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && accounts.length === 0 && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">
          No accounts yet. Create one to start recording ledger entries.
        </div>
      )}

      {!isLoading &&
        !error &&
        grouped.map((g) => (
          <div key={g.type} className="mb-6">
            <h2 className="text-sm font-semibold mb-2 capitalize text-slate-300">{g.type}</h2>
            <div className="border border-slate-800 rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-slate-900 text-slate-400 text-left">
                    <th className="px-4 py-2 font-medium">Code</th>
                    <th className="px-4 py-2 font-medium">Name</th>
                    <th className="px-4 py-2 font-medium">Status</th>
                    <th className="px-4 py-2 font-medium text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {g.items.map((a) => (
                    <tr key={a.id} className="border-t border-slate-800">
                      <td className="px-4 py-2 font-mono text-xs">{a.code}</td>
                      <td className="px-4 py-2 font-medium">{a.name}</td>
                      <td className="px-4 py-2">
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full ${
                            a.is_active ? 'bg-emerald-600/15 text-emerald-400' : 'bg-slate-800 text-slate-500'
                          }`}
                        >
                          {a.is_active ? 'Active' : 'Inactive'}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-right">
                        <button
                          onClick={() => startEdit(a)}
                          className="text-xs text-slate-400 hover:text-emerald-400 transition-colors"
                        >
                          Edit
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
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
