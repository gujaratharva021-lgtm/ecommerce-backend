import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { listDeliveryZones, createDeliveryZone, updateDeliveryZone, deleteDeliveryZone } from '../api/admin'
import type { DeliveryZone } from '../types/admin'

export default function DeliveryZones() {
  const [zones, setZones] = useState<DeliveryZone[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [city, setCity] = useState('')
  const [pincodes, setPincodes] = useState('')
  const [deliveryCharge, setDeliveryCharge] = useState('0')
  const [isCodAvailable, setIsCodAvailable] = useState(true)
  const [estimatedDays, setEstimatedDays] = useState('3')
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listDeliveryZones()
      setZones(res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load delivery zones.')
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

    if (!name.trim() || !pincodes.trim()) {
      setFormError('Name and pincodes are required.')
      return
    }

    setIsSaving(true)
    try {
      await createDeliveryZone({
        name: name.trim(),
        city: city.trim(),
        pincodes: pincodes.trim(),
        delivery_charge: parseFloat(deliveryCharge) || 0,
        is_cod_available: isCodAvailable,
        estimated_days: parseInt(estimatedDays, 10) || 3,
      })
      setName('')
      setCity('')
      setPincodes('')
      setDeliveryCharge('0')
      setIsCodAvailable(true)
      setEstimatedDays('3')
      setShowForm(false)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to create delivery zone.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleToggleStatus(z: DeliveryZone) {
    try {
      await updateDeliveryZone(z.id, { is_active: !z.is_active })
      setZones((prev) =>
        prev.map((x) => (x.id === z.id ? { ...x, is_active: !x.is_active } : x))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update zone status.')
    }
  }

  async function handleDelete(z: DeliveryZone) {
    if (!confirm('Delete delivery zone "' + z.name + '"? This cannot be undone.')) return
    try {
      await deleteDeliveryZone(z.id)
      setZones((prev) => prev.filter((x) => x.id !== z.id))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to delete delivery zone.')
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Delivery Zones</h1>
            <p className="text-sm text-slate-400 mt-1">
              Serviceable pincodes, delivery charges, and COD availability
            </p>
          </div>
          <button
            onClick={() => setShowForm((v) => !v)}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            {showForm ? 'Cancel' : '+ New Zone'}
          </button>
        </div>

        {showForm && (
          <form
            onSubmit={handleCreate}
            className="border border-slate-800 rounded-xl p-6 bg-slate-900 space-y-4 mb-6"
          >
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Zone Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Ahmedabad Local"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">City</label>
                <input
                  type="text"
                  value={city}
                  onChange={(e) => setCity(e.target.value)}
                  placeholder="e.g. Ahmedabad"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Pincodes (comma-separated)</label>
              <textarea
                value={pincodes}
                onChange={(e) => setPincodes(e.target.value)}
                placeholder="380001, 380002, 380015"
                rows={2}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500 resize-none"
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Delivery Charge (Rs)</label>
                <input
                  type="number"
                  value={deliveryCharge}
                  onChange={(e) => setDeliveryCharge(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Estimated Days</label>
                <input
                  type="number"
                  value={estimatedDays}
                  onChange={(e) => setEstimatedDays(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
              <div className="flex items-end pb-2">
                <label className="flex items-center gap-2 text-sm text-slate-400">
                  <input
                    type="checkbox"
                    checked={isCodAvailable}
                    onChange={(e) => setIsCodAvailable(e.target.checked)}
                    className="rounded border-slate-700 bg-slate-800"
                  />
                  COD Available
                </label>
              </div>
            </div>

            {formError && <p className="text-red-400 text-sm">{formError}</p>}

            <button
              type="submit"
              disabled={isSaving}
              className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium disabled:opacity-40 transition-colors"
            >
              {isSaving ? 'Creating...' : 'Create Zone'}
            </button>
          </form>
        )}

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && zones.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No delivery zones created yet.
          </div>
        )}

        {!isLoading && zones.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">City</th>
                  <th className="px-4 py-3 font-medium">Pincodes</th>
                  <th className="px-4 py-3 font-medium">Charge</th>
                  <th className="px-4 py-3 font-medium">COD</th>
                  <th className="px-4 py-3 font-medium">ETA</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {zones.map((z) => (
                  <tr key={z.id} className="border-t border-slate-800">
                    <td className="px-4 py-3 text-slate-200">{z.name}</td>
                    <td className="px-4 py-3 text-slate-400">{z.city || '-'}</td>
                    <td className="px-4 py-3 text-slate-400 max-w-xs truncate" title={z.pincodes}>
                      {z.pincodes}
                    </td>
                    <td className="px-4 py-3 text-slate-400">Rs {z.delivery_charge}</td>
                    <td className="px-4 py-3 text-slate-400">{z.is_cod_available ? 'Yes' : 'No'}</td>
                    <td className="px-4 py-3 text-slate-400">{z.estimated_days}d</td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleToggleStatus(z)}
                        className={
                          'px-2 py-1 rounded-md text-xs font-medium ' +
                          (z.is_active
                            ? 'bg-emerald-500/15 text-emerald-300'
                            : 'bg-slate-700/50 text-slate-400')
                        }
                      >
                        {z.is_active ? 'Active' : 'Inactive'}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => handleDelete(z)}
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
