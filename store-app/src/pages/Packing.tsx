import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { completePacking, getPackingTask, startPacking } from '../api/warehouse'
import type { PackingTaskResponse } from '../types/warehouse'
import StatusBadge from '../components/StatusBadge'
import { getErrorMessage } from '../utils/errors'

type Zone = 'ambient' | 'chilled' | 'frozen'

const ZONE_LABELS: Record<Zone, string> = {
  ambient: 'Ambient',
  chilled: 'Chilled',
  frozen: 'Frozen',
}

export default function Packing() {
  const { orderId } = useParams()
  const navigate = useNavigate()
  const [data, setData] = useState<PackingTaskResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isStarting, setIsStarting] = useState(false)
  const [isCompleting, setIsCompleting] = useState(false)

  const [sealNumber, setSealNumber] = useState('')
  const [qcAmbient, setQcAmbient] = useState(false)
  const [qcChilled, setQcChilled] = useState(false)
  const [qcFrozen, setQcFrozen] = useState(false)
  const [qcNotes, setQcNotes] = useState('')

  const load = useCallback(async () => {
    if (!orderId) return
    setError(null)
    try {
      const res = await getPackingTask(Number(orderId))
      setData(res)
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to load packing task.'))
    } finally {
      setIsLoading(false)
    }
  }, [orderId])

  useEffect(() => {
    load()
  }, [load])

  async function handleStart() {
    if (!orderId) return
    setIsStarting(true)
    setError(null)
    try {
      await startPacking(Number(orderId))
      await load()
    } catch (err) {
      setError(getErrorMessage(err, 'Failed to start packing.'))
    } finally {
      setIsStarting(false)
    }
  }

  async function handleComplete() {
    if (!orderId) return
    setIsCompleting(true)
    setError(null)
    try {
      await completePacking(Number(orderId), {
        seal_number: sealNumber.trim(),
        qc_ambient_ok: requiredZones.has('ambient') ? qcAmbient : undefined,
        qc_chilled_ok: requiredZones.has('chilled') ? qcChilled : undefined,
        qc_frozen_ok: requiredZones.has('frozen') ? qcFrozen : undefined,
        qc_notes: qcNotes.trim() || undefined,
      })
      await load()
    } catch (err) {
      // Missing seal, missing QC for a required zone, or double-pack
      // prevention all surface here as a 400/409/422 from the backend -
      // getErrorMessage pulls the backend's exact reason through.
      setError(getErrorMessage(err, 'Failed to complete packing.'))
    } finally {
      setIsCompleting(false)
    }
  }

  if (isLoading) return <div className="p-6 text-sm text-slate-400">Loading packing task...</div>

  if (!data) {
    return (
      <div className="p-6">
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3">
          {error ?? 'Packing task not found.'}
        </div>
      </div>
    )
  }

  const { packing_task: task, picked_items: items } = data
  const exceptions = items.filter((i) => i.status === 'unavailable' || i.status === 'short')

  // Which temperature zones are actually present among this order's items.
  // Only these get a QC checkbox - an ambient-only order never shows
  // chilled/frozen checks. Falls back to 'ambient' when a product hasn't
  // been classified yet, so nothing silently skips QC.
  const requiredZones = new Set<Zone>(
    items.map((i) => (i.product?.temperature_zone ?? 'ambient') as Zone),
  )

  const canComplete = sealNumber.trim().length > 0 &&
    (!requiredZones.has('ambient') || qcAmbient) &&
    (!requiredZones.has('chilled') || qcChilled) &&
    (!requiredZones.has('frozen') || qcFrozen)

  return (
    <div className="p-6 max-w-3xl">
      <button onClick={() => navigate('/orders')} className="text-xs text-slate-400 hover:text-slate-200 mb-3">
        &larr; Back to orders
      </button>
      <div className="flex items-center justify-between mb-1">
        <h1 className="font-display text-2xl font-semibold">Packing &mdash; Order #{task.order_id}</h1>
        <StatusBadge status={task.status} />
      </div>
      <p className="text-xs text-slate-500 mb-6">
        {task.started_at ? `Started ${new Date(task.started_at).toLocaleTimeString()}` : 'Not started yet'}
      </p>

      {error && (
        <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-sm rounded-lg px-4 py-3 mb-4">
          {error}
        </div>
      )}

      {exceptions.length > 0 && (
        <div className="border border-red-900 bg-red-950/30 text-red-300 text-sm rounded-lg px-4 py-3 mb-4">
          {exceptions.length} item(s) had picking exceptions &mdash; double-check before packing.
        </div>
      )}

      <p className="text-xs uppercase tracking-wide text-slate-500 mb-2">Picked items</p>
      <div className="space-y-2 mb-6">
        {items.map((item) => (
          <div
            key={item.id}
            className="border border-slate-800 rounded-xl bg-slate-900 p-4 flex items-center justify-between"
          >
            <div>
              <p className="font-medium text-sm">{item.product?.name ?? `Product #${item.product_id}`}</p>
              <p className="text-xs text-slate-400 mt-1">
                Needed: {item.quantity_needed} &middot; Picked: {item.quantity_picked}
              </p>
              {item.reason && <p className="text-xs text-red-400 mt-1">Reason: {item.reason}</p>}
            </div>
            <StatusBadge status={item.status} />
          </div>
        ))}
      </div>

      {task.status === 'pending' && (
        <button
          onClick={handleStart}
          disabled={isStarting}
          className="px-4 py-2 rounded-lg bg-red-500 hover:bg-red-400 text-white text-sm font-medium transition-colors disabled:opacity-50"
        >
          {isStarting ? 'Starting...' : 'Start Packing'}
        </button>
      )}

      {task.status === 'in_progress' && (
        <div className="border border-slate-800 rounded-xl bg-slate-900 p-5 space-y-4">
          <p className="text-xs uppercase tracking-wide text-slate-500">Temperature Zones</p>
          <div className="space-y-2">
            {(['ambient', 'chilled', 'frozen'] as Zone[]).map((zone) => {
              const required = requiredZones.has(zone)
              const checked = zone === 'ambient' ? qcAmbient : zone === 'chilled' ? qcChilled : qcFrozen
              const setChecked = zone === 'ambient' ? setQcAmbient : zone === 'chilled' ? setQcChilled : setQcFrozen
              return (
                <label
                  key={zone}
                  className={`flex items-center gap-2 text-sm ${required ? 'text-slate-200' : 'text-slate-600'}`}
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    disabled={!required}
                    onChange={(e) => setChecked(e.target.checked)}
                    className="accent-emerald-500"
                  />
                  {ZONE_LABELS[zone]} QC Passed
                  {!required && <span className="text-xs text-slate-600">&mdash; not required</span>}
                </label>
              )
            })}
          </div>

          <div>
            <label className="block text-xs uppercase tracking-wide text-slate-500 mb-1">Seal Number</label>
            <input
              type="text"
              value={sealNumber}
              onChange={(e) => setSealNumber(e.target.value)}
              placeholder={`SEAL-2026-${String(task.order_id).padStart(6, '0')}`}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
            />
          </div>

          <div>
            <label className="block text-xs uppercase tracking-wide text-slate-500 mb-1">QC Notes</label>
            <textarea
              value={qcNotes}
              onChange={(e) => setQcNotes(e.target.value)}
              rows={2}
              placeholder="Optional - damage, mismatch, or other notes"
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
            />
          </div>

          <button
            onClick={handleComplete}
            disabled={isCompleting || !canComplete}
            className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors disabled:opacity-50"
          >
            {isCompleting ? 'Completing...' : 'Complete Packing'}
          </button>
        </div>
      )}

      {task.status === 'completed' && (
        <div className="space-y-2">
          <span className="text-sm text-emerald-300 block">Order is ready for dispatch.</span>
          {task.seal_number && (
            <p className="text-xs text-slate-400">Seal: <span className="text-slate-200">{task.seal_number}</span></p>
          )}
          <button
            onClick={() => navigate('/orders?status=ready_for_dispatch')}
            className="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-sm font-medium transition-colors"
          >
            Back to Orders
          </button>
        </div>
      )}
    </div>
  )
}
