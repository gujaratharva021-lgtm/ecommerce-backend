import { useEffect, useState, useCallback } from 'react'
import {
  listZones,
  createZone,
  listRacks,
  createRack,
  listBins,
  createBin,
  getProductInventory,
  assignProductBin,
  deleteZone,
  deleteRack,
  deleteBin,
} from '../api/warehouse'
import type { WarehouseZone, WarehouseRack, WarehouseBin, Inventory } from '../types/warehouse'
import { getErrorMessage } from '../utils/errors'

export default function Locations() {
  const [zones, setZones] = useState<WarehouseZone[]>([])
  const [selectedZone, setSelectedZone] = useState<WarehouseZone | null>(null)
  const [racks, setRacks] = useState<WarehouseRack[]>([])
  const [selectedRack, setSelectedRack] = useState<WarehouseRack | null>(null)
  const [bins, setBins] = useState<WarehouseBin[]>([])

  const [newZoneName, setNewZoneName] = useState('')
  const [newRackName, setNewRackName] = useState('')
  const [newBinName, setNewBinName] = useState('')

  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  // Product -> bin assignment
  const [productIdInput, setProductIdInput] = useState('')
  const [lookupInventory, setLookupInventory] = useState<Inventory | null>(null)
  const [lookupError, setLookupError] = useState<string | null>(null)
  const [isLookingUp, setIsLookingUp] = useState(false)
  const [assignBinId, setAssignBinId] = useState('')
  const [isAssigning, setIsAssigning] = useState(false)

  const loadZones = useCallback(async () => {
    setError(null)
    try {
      const data = await listZones()
      setZones(data.zones)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load zones.'))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadZones()
  }, [loadZones])

  const selectZone = async (zone: WarehouseZone) => {
    setSelectedZone(zone)
    setSelectedRack(null)
    setBins([])
    try {
      const data = await listRacks(zone.id)
      setRacks(data.racks)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load racks.'))
    }
  }

  const selectRack = async (rack: WarehouseRack) => {
    setSelectedRack(rack)
    try {
      const data = await listBins(rack.id)
      setBins(data.bins)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load bins.'))
    }
  }

  const handleDeleteZone = async (zone: WarehouseZone, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm(`Delete zone "${zone.name}"? This only works if it has no racks.`)) return
    setDeletingId(`zone-${zone.id}`)
    setError(null)
    try {
      await deleteZone(zone.id)
      if (selectedZone?.id === zone.id) {
        setSelectedZone(null)
        setSelectedRack(null)
        setRacks([])
        setBins([])
      }
      await loadZones()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete zone.'))
    } finally {
      setDeletingId(null)
    }
  }

  const handleDeleteRack = async (rack: WarehouseRack, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm(`Delete rack "${rack.name}"? This only works if it has no bins.`)) return
    setDeletingId(`rack-${rack.id}`)
    setError(null)
    try {
      await deleteRack(rack.id)
      if (selectedRack?.id === rack.id) {
        setSelectedRack(null)
        setBins([])
      }
      if (selectedZone) {
        const data = await listRacks(selectedZone.id)
        setRacks(data.racks)
      }
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete rack.'))
    } finally {
      setDeletingId(null)
    }
  }

  const handleDeleteBin = async (bin: WarehouseBin, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm(`Delete bin "${bin.name}"? This only works if no product is assigned to it.`)) return
    setDeletingId(`bin-${bin.id}`)
    setError(null)
    try {
      await deleteBin(bin.id)
      if (selectedRack) {
        const data = await listBins(selectedRack.id)
        setBins(data.bins)
      }
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to delete bin.'))
    } finally {
      setDeletingId(null)
    }
  }

  const submitZone = async () => {
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

  const submitRack = async () => {
    if (!newRackName.trim() || !selectedZone) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createRack(selectedZone.id, newRackName.trim())
      setNewRackName('')
      const data = await listRacks(selectedZone.id)
      setRacks(data.racks)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create rack.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  const submitBin = async () => {
    if (!newBinName.trim() || !selectedRack) return
    setIsSubmitting(true)
    setError(null)
    try {
      await createBin(selectedRack.id, newBinName.trim())
      setNewBinName('')
      const data = await listBins(selectedRack.id)
      setBins(data.bins)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to create bin.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  const lookupProduct = async () => {
    const id = parseInt(productIdInput, 10)
    if (!id) return
    setIsLookingUp(true)
    setLookupError(null)
    setLookupInventory(null)
    try {
      const inv = await getProductInventory(id)
      setLookupInventory(inv)
      setAssignBinId(inv.bin_id ? String(inv.bin_id) : '')
    } catch (err) {
      setLookupError(getErrorMessage(err, 'No inventory found for this product in your warehouse.'))
    } finally {
      setIsLookingUp(false)
    }
  }

  const submitAssign = async () => {
    if (!lookupInventory) return
    const binId = assignBinId ? parseInt(assignBinId, 10) : null
    setIsAssigning(true)
    setLookupError(null)
    try {
      const updated = await assignProductBin(lookupInventory.product_id, binId)
      setLookupInventory(updated)
    } catch (err) {
      setLookupError(getErrorMessage(err, 'Failed to assign bin.'))
    } finally {
      setIsAssigning(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-display text-2xl font-semibold">Warehouse Locations</h1>
      </div>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {isLoading ? (
        <p className="text-sm text-slate-400">Loading zones...</p>
      ) : (
        <div className="grid grid-cols-3 gap-4 mb-8">
          {/* Zones */}
          <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden flex flex-col">
            <div className="px-4 py-2.5 bg-slate-800/50 text-xs uppercase text-slate-400 font-medium">
              Zones
            </div>
            <div className="flex-1 divide-y divide-slate-800 max-h-96 overflow-y-auto">
              {zones.length === 0 && (
                <p className="text-xs text-slate-500 px-4 py-4">No zones yet.</p>
              )}
              {zones.map((z) => (
                <div
                  key={z.id}
                  onClick={() => selectZone(z)}
                  className={`w-full flex items-center justify-between px-4 py-2.5 text-sm transition-colors cursor-pointer ${
                    selectedZone?.id === z.id
                      ? 'bg-indigo-500/15 text-indigo-300'
                      : 'text-slate-300 hover:bg-slate-800/50'
                  }`}
                >
                  <span>{z.name}</span>
                  <button
                    onClick={(e) => handleDeleteZone(z, e)}
                    disabled={deletingId === `zone-${z.id}`}
                    className="text-slate-600 hover:text-rose-400 disabled:opacity-40 text-xs px-1.5"
                    title="Delete zone (only if empty)"
                  >
                    {deletingId === `zone-${z.id}` ? '...' : '\u2715'}
                  </button>
                </div>
              ))}
            </div>
            <div className="p-3 border-t border-slate-800 flex gap-1.5">
              <input
                value={newZoneName}
                onChange={(e) => setNewZoneName(e.target.value)}
                placeholder="New zone name"
                className="flex-1 text-xs bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-200 focus:outline-none focus:border-indigo-500"
              />
              <button
                onClick={submitZone}
                disabled={isSubmitting || !newZoneName.trim()}
                className="text-xs px-3 py-1.5 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-40"
              >
                Add
              </button>
            </div>
          </div>

          {/* Racks */}
          <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden flex flex-col">
            <div className="px-4 py-2.5 bg-slate-800/50 text-xs uppercase text-slate-400 font-medium">
              Racks {selectedZone ? `— ${selectedZone.name}` : ''}
            </div>
            <div className="flex-1 divide-y divide-slate-800 max-h-96 overflow-y-auto">
              {!selectedZone && (
                <p className="text-xs text-slate-500 px-4 py-4">Select a zone to view its racks.</p>
              )}
              {selectedZone && racks.length === 0 && (
                <p className="text-xs text-slate-500 px-4 py-4">No racks yet.</p>
              )}
              {racks.map((r) => (
                <div
                  key={r.id}
                  onClick={() => selectRack(r)}
                  className={`w-full flex items-center justify-between px-4 py-2.5 text-sm transition-colors cursor-pointer ${
                    selectedRack?.id === r.id
                      ? 'bg-indigo-500/15 text-indigo-300'
                      : 'text-slate-300 hover:bg-slate-800/50'
                  }`}
                >
                  <span>{r.name}</span>
                  <button
                    onClick={(e) => handleDeleteRack(r, e)}
                    disabled={deletingId === `rack-${r.id}`}
                    className="text-slate-600 hover:text-rose-400 disabled:opacity-40 text-xs px-1.5"
                    title="Delete rack (only if empty)"
                  >
                    {deletingId === `rack-${r.id}` ? '...' : '\u2715'}
                  </button>
                </div>
              ))}
            </div>
            {selectedZone && (
              <div className="p-3 border-t border-slate-800 flex gap-1.5">
                <input
                  value={newRackName}
                  onChange={(e) => setNewRackName(e.target.value)}
                  placeholder="New rack name"
                  className="flex-1 text-xs bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-200 focus:outline-none focus:border-indigo-500"
                />
                <button
                  onClick={submitRack}
                  disabled={isSubmitting || !newRackName.trim()}
                  className="text-xs px-3 py-1.5 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-40"
                >
                  Add
                </button>
              </div>
            )}
          </div>

          {/* Bins */}
          <div className="border border-slate-800 rounded-xl bg-slate-900 overflow-hidden flex flex-col">
            <div className="px-4 py-2.5 bg-slate-800/50 text-xs uppercase text-slate-400 font-medium">
              Bins {selectedRack ? `— ${selectedRack.name}` : ''}
            </div>
            <div className="flex-1 divide-y divide-slate-800 max-h-96 overflow-y-auto">
              {!selectedRack && (
                <p className="text-xs text-slate-500 px-4 py-4">Select a rack to view its bins.</p>
              )}
              {selectedRack && bins.length === 0 && (
                <p className="text-xs text-slate-500 px-4 py-4">No bins yet.</p>
              )}
              {bins.map((b) => (
                <div key={b.id} className="px-4 py-2.5 text-sm text-slate-300 flex items-center justify-between">
                  <span>
                    {b.name}
                    <span className="text-slate-600 ml-2 text-xs">#{b.id}</span>
                  </span>
                  <button
                    onClick={(e) => handleDeleteBin(b, e)}
                    disabled={deletingId === `bin-${b.id}`}
                    className="text-slate-600 hover:text-rose-400 disabled:opacity-40 text-xs px-1.5"
                    title="Delete bin (only if unassigned)"
                  >
                    {deletingId === `bin-${b.id}` ? '...' : '\u2715'}
                  </button>
                </div>
              ))}
            </div>
            {selectedRack && (
              <div className="p-3 border-t border-slate-800 flex gap-1.5">
                <input
                  value={newBinName}
                  onChange={(e) => setNewBinName(e.target.value)}
                  placeholder="New bin name"
                  className="flex-1 text-xs bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-200 focus:outline-none focus:border-indigo-500"
                />
                <button
                  onClick={submitBin}
                  disabled={isSubmitting || !newBinName.trim()}
                  className="text-xs px-3 py-1.5 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-40"
                >
                  Add
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Assign product to bin */}
      <div className="border border-slate-800 rounded-xl bg-slate-900 p-5 max-w-2xl">
        <h2 className="text-sm font-semibold mb-3">Assign Product to Bin</h2>
        <div className="flex gap-2 mb-3">
          <input
            value={productIdInput}
            onChange={(e) => setProductIdInput(e.target.value)}
            placeholder="Product ID"
            className="flex-1 text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
          />
          <button
            onClick={lookupProduct}
            disabled={isLookingUp || !productIdInput}
            className="text-xs px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40"
          >
            {isLookingUp ? 'Looking up...' : 'Look up'}
          </button>
        </div>

        {lookupError && (
          <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-xs rounded-lg px-3 py-2 mb-3">
            {lookupError}
          </div>
        )}

        {lookupInventory && (
          <div className="border border-slate-800 rounded-lg p-3 bg-slate-950/50">
            <p className="text-sm font-medium">{lookupInventory.product?.name ?? `Product #${lookupInventory.product_id}`}</p>
            <p className="text-xs text-slate-500 mb-3">
              Stock: {lookupInventory.stock} &middot; Current bin:{' '}
              {lookupInventory.bin
                ? `${lookupInventory.bin.rack?.zone?.name ?? ''} / ${lookupInventory.bin.rack?.name ?? ''} / ${lookupInventory.bin.name}`
                : 'Unassigned'}
            </p>
            <div className="flex gap-2">
              <input
                value={assignBinId}
                onChange={(e) => setAssignBinId(e.target.value)}
                placeholder="Bin ID (leave empty to clear)"
                className="flex-1 text-sm bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-indigo-500"
              />
              <button
                onClick={submitAssign}
                disabled={isAssigning}
                className="text-xs px-4 py-2 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 disabled:opacity-50"
              >
                {isAssigning ? 'Saving...' : 'Assign'}
              </button>
            </div>
            <p className="text-xs text-slate-600 mt-2">
              Tip: select a bin above to see its ID next to its name.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
