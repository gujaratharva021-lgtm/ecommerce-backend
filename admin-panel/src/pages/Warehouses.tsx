import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import Modal from '../components/Modal'
import ServiceAreaModal from '../components/ServiceAreaModal'
import {
  listWarehouses,
  createWarehouse,
  updateWarehouse,
  deleteWarehouse,
} from '../api/admin'
import type { Warehouse } from '../types/admin'

const emptyForm = {
  name: '',
  city: '',
  address: '',
  lat: '',
  lng: '',
  service_radius_km: '5',
  is_active: true,
}

export default function Warehouses() {
  const [warehouses, setWarehouses] = useState<Warehouse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showCreate, setShowCreate] = useState(false)
  const [editingWarehouse, setEditingWarehouse] = useState<Warehouse | null>(null)
  const [serviceAreaWarehouse, setServiceAreaWarehouse] = useState<Warehouse | null>(null)

  const [form, setForm] = useState(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listWarehouses()
      setWarehouses(res.warehouses ?? res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load warehouses.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  function openCreate() {
    setForm(emptyForm)
    setFormError(null)
    setEditingWarehouse(null)
    setShowCreate(true)
  }

  function openEdit(w: Warehouse) {
    setForm({
      name: w.name,
      city: w.city,
      address: w.address ?? '',
      lat: String(w.lat),
      lng: String(w.lng),
      service_radius_km: String(w.service_radius_km ?? 5),
      is_active: w.is_active ?? true,
    })
    setFormError(null)
    setEditingWarehouse(w)
    setShowCreate(true)
  }

  function closeModal() {
    setShowCreate(false)
    setEditingWarehouse(null)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)

    if (!form.name.trim() || !form.city.trim() || !form.lat || !form.lng) {
      setFormError('Name, city, latitude, and longitude are required.')
      return
    }

    const lat = parseFloat(form.lat)
    const lng = parseFloat(form.lng)
    if (Number.isNaN(lat) || Number.isNaN(lng)) {
      setFormError('Latitude and longitude must be valid numbers.')
      return
    }

    setIsSaving(true)
    try {
      const payload = {
        name: form.name.trim(),
        city: form.city.trim(),
        address: form.address.trim(),
        lat,
        lng,
        service_radius_km: form.service_radius_km ? parseFloat(form.service_radius_km) : 5,
        is_active: form.is_active,
      }
      if (editingWarehouse) {
        await updateWarehouse(editingWarehouse.id, payload)
      } else {
        await createWarehouse(payload)
      }
      closeModal()
      setForm(emptyForm)
      load()
    } catch (err: any) {
      setFormError(
        err.response?.data?.error ??
          `Failed to ${editingWarehouse ? 'update' : 'create'} warehouse.`
      )
    } finally {
      setIsSaving(false)
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Delete this warehouse? This cannot be undone.')) return
    try {
      await deleteWarehouse(id)
      setWarehouses((prev) => prev.filter((w) => w.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to delete warehouse.')
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Warehouses</h1>
            <p className="text-sm text-slate-400 mt-1">
              {warehouses.length} warehouse{warehouses.length !== 1 ? 's' : ''}
            </p>
          </div>
          <button
            onClick={openCreate}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            + Add warehouse
          </button>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && warehouses.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No warehouses yet. Add your first one to get started.
          </div>
        )}

        {!isLoading && warehouses.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">City</th>
                  <th className="px-4 py-3 font-medium">Lat, Lng</th>
                  <th className="px-4 py-3 font-medium">Radius (km)</th>
                  <th className="px-4 py-3 font-medium">Service Area</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {warehouses.map((w) => (
                  <tr key={w.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">{w.name}</td>
                    <td className="px-4 py-3">{w.city}</td>
                    <td className="px-4 py-3 text-slate-400">
                      {w.lat?.toFixed(4)}, {w.lng?.toFixed(4)}
                    </td>
                    <td className="px-4 py-3">{w.service_radius_km ?? 5}</td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => setServiceAreaWarehouse(w)}
                        className={`text-xs px-2 py-1 rounded-full transition-colors ${
                          w.service_area
                            ? 'bg-emerald-500/15 text-emerald-400 hover:bg-emerald-500/25'
                            : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                        }`}
                      >
                        {w.service_area ? 'Set \u2713 (edit)' : 'Set polygon'}
                      </button>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs px-2 py-1 rounded-full ${
                          w.is_active
                            ? 'bg-emerald-500/15 text-emerald-400'
                            : 'bg-slate-700 text-slate-300'
                        }`}
                      >
                        {w.is_active ? 'active' : 'inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right space-x-3">
                      <button
                        onClick={() => openEdit(w)}
                        className="text-indigo-400 hover:text-indigo-300 text-xs"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(w.id)}
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

      {showCreate && (
        <Modal
          title={editingWarehouse ? 'Edit warehouse' : 'Add warehouse'}
          onClose={closeModal}
        >
          <form onSubmit={handleSubmit} className="space-y-3">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Name</label>
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">City</label>
              <input
                value={form.city}
                onChange={(e) => setForm({ ...form, city: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Address</label>
              <input
                value={form.address}
                onChange={(e) => setForm({ ...form, address: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Latitude</label>
                <input
                  type="number"
                  step="any"
                  value={form.lat}
                  onChange={(e) => setForm({ ...form, lat: e.target.value })}
                  placeholder="e.g. 23.0225"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="text-xs text-slate-400 block mb-1">Longitude</label>
                <input
                  type="number"
                  step="any"
                  value={form.lng}
                  onChange={(e) => setForm({ ...form, lng: e.target.value })}
                  placeholder="e.g. 72.5714"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Service Radius (km)</label>
              <input
                type="number"
                step="any"
                value={form.service_radius_km}
                onChange={(e) => setForm({ ...form, service_radius_km: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                id="w_is_active"
                type="checkbox"
                checked={form.is_active}
                onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                className="rounded border-slate-700 bg-slate-800"
              />
              <label htmlFor="w_is_active" className="text-sm text-slate-300">
                Active
              </label>
            </div>

            {formError && <p className="text-red-400 text-xs">{formError}</p>}

            <button
              type="submit"
              disabled={isSaving}
              className="w-full py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors mt-2"
            >
              {isSaving
                ? 'Saving...'
                : editingWarehouse
                ? 'Save changes'
                : 'Add warehouse'}
            </button>
          </form>
        </Modal>
      )}

      {serviceAreaWarehouse && (
        <ServiceAreaModal
          warehouse={serviceAreaWarehouse}
          onClose={() => setServiceAreaWarehouse(null)}
          onSaved={load}
        />
      )}
    </Layout>
  )
}
