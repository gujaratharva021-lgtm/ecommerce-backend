import { useCallback, useEffect, useState } from 'react'
import { listZones, createZone, deleteZone, listRacks, createRack, deleteRack, listBins, createBin, deleteBin } from '../api/warehouse'
import type { WarehouseZone, WarehouseRack, WarehouseBin } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

export default function Locations() {
  const [zones, setZones] = useState<WarehouseZone[]>([])
  const [racks, setRacks] = useState<WarehouseRack[]>([])
  const [bins, setBins] = useState<WarehouseBin[]>([])

  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null)
  const [selectedRackId, setSelectedRackId] = useState<number | null>(null)

  const [newZoneName, setNewZoneName] = useState('')
  const [newRackName, setNewRackName] = useState('')
  const [newBinName, setNewBinName] = useState('')

  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const loadZones = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listZones()
      setZones(res.zones ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load zones.'))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadZones()
  }, [loadZones])

  useEffect(() => {
    if (selectedZoneId == null) {
      setRacks([])
      setSelectedRackId(null)
      return
    }
    listRacks(selectedZoneId)
      .then((res) => setRacks(res.racks ?? []))
      .catch((err) => setError(getErrorMessage(err, 'Failed to load racks.')))
  }, [selectedZoneId])

  useEffect(() => {
    if (selectedRackId == null) {
      setBins([])
      return
    }
    listBins(selectedRackId)
      .then((res) => setBins(res.bins ?? []))
      .catch((err) => setError(getErrorMessage(err, 'Failed to load bins.')))
  }, [selectedRackId])

  async function handleCreateZone() {
    if (!newZoneName.trim()) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createZone(newZoneName.trim())
      setNewZoneName('')
      await loadZones()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create zone.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleCreateRack() {
    if (!newRackName.trim() || selectedZoneId == null) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createRack(selectedZoneId, newRackName.trim())
      setNewRackName('')
      const res = await listRacks(selectedZoneId)
      setRacks(res.racks ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create rack.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleCreateBin() {
    if (!newBinName.trim() || selectedRackId == null) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createBin(selectedRackId, newBinName.trim())
      setNewBinName('')
      const res = await listBins(selectedRackId)
      setBins(res.bins ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create bin.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleDeleteZone(id: number) {
    setError(null)
    try {
      await deleteZone(id)
      if (selectedZoneId === id) setSelectedZoneId(null)
      await loadZones()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete zone.'))
    }
  }

  async function handleDeleteRack(id: number) {
    if (selectedZoneId == null) return
    setError(null)
    try {
      await deleteRack(id)
      if (selectedRackId === id) setSelectedRackId(null)
      const res = await listRacks(selectedZoneId)
      setRacks(res.racks ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete rack.'))
    }
  }

  async function handleDeleteBin(id: number) {
    if (selectedRackId == null) return
    setError(null)
    try {
      await deleteBin(id)
      const res = await listBins(selectedRackId)
      setBins(res.bins ?? [])
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete bin.'))
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-1">
        <h1 className="font-display text-2xl font-semibold">Locations</h1>
      </div>
      <p className="text-sm text-slate-400 mb-5">Manage zones, racks and bins for your warehouse.</p>

      {error && (
        <div className="mb-4 px-4 py-2 rounded-lg bg-red-500/10 border border-red-500/30 text-red-300 text-sm">
          {error}
        </div>
      )}

      {isLoading ? (
        <p className="text-sm text-slate-400">Loading...</p>
      ) : (
        <div className="grid grid-cols-3 gap-4">
          {/* Zones column */}
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
            <h2 className="text-sm font-semibold text-slate-200 mb-3">Zones</h2>
            <div className="space-y-1 mb-3">
              {zones.length === 0 && <p className="text-xs text-slate-500">No zones yet.</p>}
              {zones.map((zone) => (
                <div
                  key={zone.id}
                  onClick={() => { setSelectedZoneId(zone.id); setSelectedRackId(null) }}
                  className={`flex items-center justify-between pl-3 pr-2 py-2 text-sm rounded-lg cursor-pointer border-l-2 transition-colors ${
                    selectedZoneId === zone.id
                      ? 'border-red-500 bg-red-500/10 text-red-300 font-medium'
                      : 'border-transparent text-slate-300 hover:bg-slate-800'
                  }`}
                >
                  <span>{zone.name}</span>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDeleteZone(zone.id) }}
                    className="text-xs text-slate-500 hover:text-red-400"
                  >
                    Delete
                  </button>
                </div>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                value={newZoneName}
                onChange={(e) => setNewZoneName(e.target.value)}
                placeholder="New zone name"
                className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
              />
              <button
                disabled={isSubmitting || !newZoneName.trim()}
                onClick={handleCreateZone}
                className="px-3 py-1.5 rounded-lg bg-red-500 hover:bg-red-400 text-white text-xs font-medium disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </div>

          {/* Racks column */}
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
            <h2 className="text-sm font-semibold text-slate-200 mb-3">Racks</h2>
            {selectedZoneId == null ? (
              <p className="text-xs text-slate-500">Select a zone to view racks.</p>
            ) : (
              <>
                <div className="space-y-1 mb-3">
                  {racks.length === 0 && <p className="text-xs text-slate-500">No racks yet.</p>}
                  {racks.map((rack) => (
                    <div
                      key={rack.id}
                      onClick={() => setSelectedRackId(rack.id)}
                      className={`flex items-center justify-between pl-3 pr-2 py-2 text-sm rounded-lg cursor-pointer border-l-2 transition-colors ${
                        selectedRackId === rack.id
                          ? 'border-red-500 bg-red-500/10 text-red-300 font-medium'
                          : 'border-transparent text-slate-300 hover:bg-slate-800'
                      }`}
                    >
                      <span>{rack.name}</span>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDeleteRack(rack.id) }}
                        className="text-xs text-slate-500 hover:text-red-400"
                      >
                        Delete
                      </button>
                    </div>
                  ))}
                </div>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newRackName}
                    onChange={(e) => setNewRackName(e.target.value)}
                    placeholder="New rack name"
                    className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
                  />
                  <button
                    disabled={isSubmitting || !newRackName.trim()}
                    onClick={handleCreateRack}
                    className="px-3 py-1.5 rounded-lg bg-red-500 hover:bg-red-400 text-white text-xs font-medium disabled:opacity-50"
                  >
                    Add
                  </button>
                </div>
              </>
            )}
          </div>

          {/* Bins column */}
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
            <h2 className="text-sm font-semibold text-slate-200 mb-3">Bins</h2>
            {selectedRackId == null ? (
              <p className="text-xs text-slate-500">Select a rack to view bins.</p>
            ) : (
              <>
                <div className="space-y-1 mb-3">
                  {bins.length === 0 && <p className="text-xs text-slate-500">No bins yet.</p>}
                  {bins.map((bin) => (
                    <div
                      key={bin.id}
                      className="flex items-center justify-between pl-3 pr-2 py-2 text-sm rounded-lg text-slate-300 border-l-2 border-transparent"
                    >
                      <span>{bin.name}</span>
                      <button
                        onClick={() => handleDeleteBin(bin.id)}
                        className="text-xs text-slate-500 hover:text-red-400"
                      >
                        Delete
                      </button>
                    </div>
                  ))}
                </div>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newBinName}
                    onChange={(e) => setNewBinName(e.target.value)}
                    placeholder="New bin name"
                    className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs"
                  />
                  <button
                    disabled={isSubmitting || !newBinName.trim()}
                    onClick={handleCreateBin}
                    className="px-3 py-1.5 rounded-lg bg-red-500 hover:bg-red-400 text-white text-xs font-medium disabled:opacity-50"
                  >
                    Add
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
