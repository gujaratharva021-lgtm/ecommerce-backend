import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { listBanners, createBanner, updateBanner, deleteBanner } from '../api/admin'
import type { Banner } from '../types/admin'

export default function Banners() {
  const [banners, setBanners] = useState<Banner[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [imageUrl, setImageUrl] = useState('')
  const [title, setTitle] = useState('')
  const [linkType, setLinkType] = useState('none')
  const [linkValue, setLinkValue] = useState('')
  const [displayOrder, setDisplayOrder] = useState('0')
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listBanners()
      setBanners(res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load banners.')
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

    if (!imageUrl.trim()) {
      setFormError('Image URL is required.')
      return
    }

    setIsSaving(true)
    try {
      await createBanner({
        image_url: imageUrl.trim(),
        title: title.trim(),
        link_type: linkType,
        link_value: linkValue.trim(),
        display_order: parseInt(displayOrder, 10) || 0,
      })
      setImageUrl('')
      setTitle('')
      setLinkType('none')
      setLinkValue('')
      setDisplayOrder('0')
      setShowForm(false)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to create banner.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleToggleStatus(b: Banner) {
    try {
      await updateBanner(b.id, { is_active: !b.is_active })
      setBanners((prev) =>
        prev.map((x) => (x.id === b.id ? { ...x, is_active: !x.is_active } : x))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update banner status.')
    }
  }

  async function handleDelete(b: Banner) {
    if (!confirm('Delete banner "' + (b.title || b.image_url) + '"? This cannot be undone.')) return
    try {
      await deleteBanner(b.id)
      setBanners((prev) => prev.filter((x) => x.id !== b.id))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to delete banner.')
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Banners</h1>
            <p className="text-sm text-slate-400 mt-1">
              Homepage carousel images shown to customers
            </p>
          </div>
          <button
            onClick={() => setShowForm((v) => !v)}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            {showForm ? 'Cancel' : '+ New Banner'}
          </button>
        </div>

        {showForm && (
          <form
            onSubmit={handleCreate}
            className="border border-slate-800 rounded-xl p-6 bg-slate-900 space-y-4 mb-6"
          >
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

            <div>
              <label className="block text-sm text-slate-400 mb-1">Title (optional)</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g. Summer Sale Banner"
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Link Type</label>
                <select
                  value={linkType}
                  onChange={(e) => setLinkType(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                >
                  <option value="none">None</option>
                  <option value="product">Product</option>
                  <option value="category">Category</option>
                  <option value="url">URL</option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Link Value</label>
                <input
                  type="text"
                  value={linkValue}
                  onChange={(e) => setLinkValue(e.target.value)}
                  placeholder="product/category id or URL"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Display Order</label>
                <input
                  type="number"
                  value={displayOrder}
                  onChange={(e) => setDisplayOrder(e.target.value)}
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
              {isSaving ? 'Creating...' : 'Create Banner'}
            </button>
          </form>
        )}

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && banners.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No banners created yet.
          </div>
        )}

        {!isLoading && banners.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Preview</th>
                  <th className="px-4 py-3 font-medium">Title</th>
                  <th className="px-4 py-3 font-medium">Link</th>
                  <th className="px-4 py-3 font-medium">Order</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {banners.map((b) => (
                  <tr key={b.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">
                      <img
                        src={b.image_url}
                        alt={b.title}
                        className="h-10 w-20 object-cover rounded-md bg-slate-800"
                        onError={(e) => {
                          ;(e.target as HTMLImageElement).style.display = 'none'
                        }}
                      />
                    </td>
                    <td className="px-4 py-3 text-slate-200">{b.title || '-'}</td>
                    <td className="px-4 py-3 text-slate-400">
                      {b.link_type !== 'none' ? `${b.link_type}: ${b.link_value}` : '-'}
                    </td>
                    <td className="px-4 py-3 text-slate-400">{b.display_order}</td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleToggleStatus(b)}
                        className={
                          'px-2 py-1 rounded-md text-xs font-medium ' +
                          (b.is_active
                            ? 'bg-emerald-500/15 text-emerald-300'
                            : 'bg-slate-700/50 text-slate-400')
                        }
                      >
                        {b.is_active ? 'Active' : 'Inactive'}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => handleDelete(b)}
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
