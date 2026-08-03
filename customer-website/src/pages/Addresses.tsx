import { useEffect, useState } from 'react'
import { listAddresses, createAddress, deleteAddress, setDefaultAddress } from '../api/addresses'
import type { Address, AddressRequest } from '../types'

const emptyForm: AddressRequest = {
  label: '',
  full_name: '',
  phone: '',
  line1: '',
  line2: '',
  city: '',
  state: '',
  pincode: '',
  is_default: false,
}

export default function Addresses() {
  const [addresses, setAddresses] = useState<Address[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<AddressRequest>(emptyForm)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function load() {
    setIsLoading(true)
    listAddresses()
      .then(setAddresses)
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      await createAddress(form)
      setForm(emptyForm)
      setShowForm(false)
      load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to save address.')
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleDelete(id: number) {
    await deleteAddress(id)
    load()
  }

  async function handleSetDefault(id: number) {
    await setDefaultAddress(id)
    load()
  }

  return (
    <div className="max-w-2xl mx-auto px-6 py-10">
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-display text-3xl font-600">Your addresses</h1>
        {!showForm && (
          <button
            onClick={() => setShowForm(true)}
            className="text-sm font-medium bg-ink text-paper px-4 py-2 rounded-lg hover:bg-marigold transition-colors"
          >
            + Add address
          </button>
        )}
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="border border-line rounded-xl p-5 mb-6 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <input
              placeholder="Label (e.g. Home)"
              value={form.label}
              onChange={(e) => setForm({ ...form, label: e.target.value })}
              className="border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
            />
            <input
              placeholder="Full name"
              value={form.full_name}
              onChange={(e) => setForm({ ...form, full_name: e.target.value })}
              className="border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
              required
            />
          </div>
          <input
            placeholder="Phone (10 digits)"
            value={form.phone}
            onChange={(e) => setForm({ ...form, phone: e.target.value.replace(/\D/g, '').slice(0, 10) })}
            className="w-full border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
            required
          />
          <input
            placeholder="Address line 1"
            value={form.line1}
            onChange={(e) => setForm({ ...form, line1: e.target.value })}
            className="w-full border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
            required
          />
          <input
            placeholder="Address line 2 (optional)"
            value={form.line2}
            onChange={(e) => setForm({ ...form, line2: e.target.value })}
            className="w-full border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
          />
          <div className="grid grid-cols-3 gap-3">
            <input
              placeholder="City"
              value={form.city}
              onChange={(e) => setForm({ ...form, city: e.target.value })}
              className="border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
              required
            />
            <input
              placeholder="State"
              value={form.state}
              onChange={(e) => setForm({ ...form, state: e.target.value })}
              className="border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
              required
            />
            <input
              placeholder="Pincode"
              value={form.pincode}
              onChange={(e) => setForm({ ...form, pincode: e.target.value.replace(/\D/g, '').slice(0, 6) })}
              className="border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
              required
            />
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={form.is_default}
              onChange={(e) => setForm({ ...form, is_default: e.target.checked })}
            />
            Set as default
          </label>

          {error && <p className="text-clay text-sm">{error}</p>}

          <div className="flex gap-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="bg-ink text-paper text-sm font-medium px-4 py-2 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
            >
              {isSubmitting ? 'Saving...' : 'Save address'}
            </button>
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="text-sm px-4 py-2 rounded-lg border border-line hover:border-ink"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {isLoading ? (
        <p className="text-ink/50">Loading...</p>
      ) : addresses.length === 0 ? (
        <p className="text-ink/50 text-sm">No saved addresses yet.</p>
      ) : (
        <div className="space-y-3">
          {addresses.map((a) => (
            <div key={a.id} className="border border-line rounded-xl p-4 flex items-start justify-between">
              <div>
                <p className="font-medium">
                  {a.label || 'Address'}{' '}
                  {a.is_default && (
                    <span className="text-xs font-mono text-leaf ml-1">(default)</span>
                  )}
                </p>
                <p className="text-sm text-ink/60 mt-1">
                  {a.full_name} · {a.phone}
                </p>
                <p className="text-sm text-ink/60">
                  {a.line1}, {a.line2 ? `${a.line2}, ` : ''}
                  {a.city}, {a.state} - {a.pincode}
                </p>
              </div>
              <div className="flex flex-col gap-1 items-end shrink-0 ml-4">
                {!a.is_default && (
                  <button
                    onClick={() => handleSetDefault(a.id)}
                    className="text-xs text-marigold hover:underline"
                  >
                    Set default
                  </button>
                )}
                <button
                  onClick={() => handleDelete(a.id)}
                  className="text-xs text-clay hover:underline"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
