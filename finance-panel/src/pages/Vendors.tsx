import { useEffect, useState } from 'react'
import { getVendors, createVendor, deleteVendor } from '../api/finance'
import type { Vendor, VendorRequest } from '../types/finance'

const emptyForm: VendorRequest = { name: '', contact_name: '', phone: '', email: '', gstin: '', address: '' }

export default function Vendors() {
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<VendorRequest>(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  function load() {
    setIsLoading(true)
    setError(null)
    getVendors()
      .then((res) => setVendors(res.vendors ?? []))
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load vendors.'))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name.trim()) {
      setFormError('Name is required.')
      return
    }
    setIsSaving(true)
    setFormError(null)
    try {
      await createVendor(form)
      setForm(emptyForm)
      setShowForm(false)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Could not create vendor.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`Delete vendor "${name}"? This cannot be undone.`)) return
    try {
      await deleteVendor(id)
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Could not delete vendor.')
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Vendors</h1>
          <p className="text-sm text-slate-500">Suppliers the business buys from.</p>
        </div>
        <button
          onClick={() => setShowForm((s) => !s)}
          className="text-sm bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg transition-colors"
        >
          {showForm ? 'Cancel' : '+ New Vendor'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="border border-slate-800 rounded-xl p-5 mb-6 max-w-2xl">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <Field label="Name *">
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Contact Name">
              <input
                value={form.contact_name}
                onChange={(e) => setForm({ ...form, contact_name: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Phone">
              <input
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Email">
              <input
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="GSTIN">
              <input
                value={form.gstin}
                onChange={(e) => setForm({ ...form, gstin: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </Field>
            <Field label="Address">
              <input
                value={form.address}
                onChange={(e) => setForm({ ...form, address: e.target.value })}
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
            {isSaving ? 'Saving...' : 'Create Vendor'}
          </button>
        </form>
      )}

      {isLoading && <p className="text-sm text-slate-500">Loading vendors...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && (
        <div className="border border-slate-800 rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-900 text-slate-400 text-left">
                <th className="px-4 py-2 font-medium">Name</th>
                <th className="px-4 py-2 font-medium">Contact</th>
                <th className="px-4 py-2 font-medium">Phone</th>
                <th className="px-4 py-2 font-medium">GSTIN</th>
                <th className="px-4 py-2 font-medium">Status</th>
                <th className="px-4 py-2 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {vendors.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-6 text-center text-slate-500">
                    No vendors yet.
                  </td>
                </tr>
              )}
              {vendors.map((v) => (
                <tr key={v.id} className="border-t border-slate-800">
                  <td className="px-4 py-2 font-medium">{v.name}</td>
                  <td className="px-4 py-2 text-slate-400">{v.contact_name || '—'}</td>
                  <td className="px-4 py-2 text-slate-400">{v.phone || '—'}</td>
                  <td className="px-4 py-2 text-slate-400">{v.gstin || '—'}</td>
                  <td className="px-4 py-2">
                    <span
                      className={`text-xs px-2 py-0.5 rounded-full ${
                        v.is_active ? 'bg-emerald-600/15 text-emerald-400' : 'bg-slate-800 text-slate-500'
                      }`}
                    >
                      {v.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={() => handleDelete(v.id, v.name)}
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
