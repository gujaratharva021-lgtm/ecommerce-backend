import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { listOffers, createOffer, updateOfferStatus, deleteOffer } from '../api/admin'
import type { Offer } from '../types/admin'

export default function Offers() {
  const [offers, setOffers] = useState<Offer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [imageUrl, setImageUrl] = useState('')
  const [discountText, setDiscountText] = useState('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listOffers()
      setOffers(res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load offers.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)

    if (!title.trim() || !startDate || !endDate) {
      setFormError('Title, start date, and end date are required.')
      return
    }
    if (new Date(endDate) < new Date(startDate)) {
      setFormError('End date must be after start date.')
      return
    }

    setIsSaving(true)
    try {
      await createOffer({
        title: title.trim(),
        description: description.trim(),
        image_url: imageUrl.trim(),
        discount_text: discountText.trim(),
        start_date: startDate,
        end_date: endDate,
      })
      setTitle('')
      setDescription('')
      setImageUrl('')
      setDiscountText('')
      setStartDate('')
      setEndDate('')
      setShowForm(false)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to create offer.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleToggleStatus(o: Offer) {
    try {
      await updateOfferStatus(o.id, !o.is_active)
      setOffers((prev) =>
        prev.map((x) => (x.id === o.id ? { ...x, is_active: !x.is_active } : x))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update offer status.')
    }
  }

  async function handleDelete(o: Offer) {
    if (!confirm('Delete offer "' + o.title + '"? This cannot be undone.')) return
    try {
      await deleteOffer(o.id)
      setOffers((prev) => prev.filter((x) => x.id !== o.id))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to delete offer.')
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Offers & Promotions</h1>
            <p className="text-sm text-slate-400 mt-1">
              Homepage and category promo campaigns shown to customers
            </p>
          </div>
          <button
            onClick={() => setShowForm((v) => !v)}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            {showForm ? 'Cancel' : '+ New Offer'}
          </button>
        </div>

        {showForm && (
          <form
            onSubmit={handleCreate}
            className="border border-slate-800 rounded-xl p-6 bg-slate-900 space-y-4 mb-6"
          >
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Title</label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="e.g. Diwali Mega Sale"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Discount Text</label>
                <input
                  type="text"
                  value={discountText}
                  onChange={(e) => setDiscountText(e.target.value)}
                  placeholder="e.g. Up to 30% off"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Description</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={2}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500 resize-none"
              />
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Image URL</label>
              <input
                type="text"
                value={imageUrl}
                onChange={(e) => setImageUrl(e.target.value)}
                placeholder="https://..."
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Start Date</label>
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">End Date</label>
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
            </div>

            {formError && <p className="text-red-400 text-sm">{formError}</p>}

            <button
              type="submit"
              disabled={isSaving}
              className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium disabled:opacity-40 transition-colors"
            >
              {isSaving ? 'Creating...' : 'Create Offer'}
            </button>
          </form>
        )}

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && offers.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No offers created yet.
          </div>
        )}

        {!isLoading && offers.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Title</th>
                  <th className="px-4 py-3 font-medium">Discount</th>
                  <th className="px-4 py-3 font-medium">Start</th>
                  <th className="px-4 py-3 font-medium">End</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {offers.map((o) => (
                  <tr key={o.id} className="border-t border-slate-800">
                    <td className="px-4 py-3 text-slate-200">{o.title}</td>
                    <td className="px-4 py-3 text-slate-400">{o.discount_text || '-'}</td>
                    <td className="px-4 py-3 text-slate-400">
                      {new Date(o.start_date).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 text-slate-400">
                      {new Date(o.end_date).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleToggleStatus(o)}
                        className={
                          'px-2 py-1 rounded-md text-xs font-medium ' +
                          (o.is_active
                            ? 'bg-emerald-500/15 text-emerald-300'
                            : 'bg-slate-700/50 text-slate-400')
                        }
                      >
                        {o.is_active ? 'Active' : 'Inactive'}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => handleDelete(o)}
                        className="text-red-400 hover:text-red-300 text-xs"
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
    </Layout>
  )
}
