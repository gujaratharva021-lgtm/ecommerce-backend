import { useEffect, useRef, useState } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-draw'
import 'leaflet-draw/dist/leaflet.draw.css'
import { setWarehouseServiceArea } from '../api/admin'
import type { Warehouse } from '../types/admin'

// Leaflet's default marker icons reference image files that Vite doesn't
// resolve automatically from the package. Point them at CDN copies so the
// warehouse marker actually renders instead of a broken image.
delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
})

export default function ServiceAreaModal({
  warehouse,
  onClose,
  onSaved,
}: {
  warehouse: Warehouse
  onClose: () => void
  onSaved: () => void
}) {
  const mapRef = useRef<HTMLDivElement | null>(null)
  const mapInstance = useRef<L.Map | null>(null)
  const drawnItems = useRef<L.FeatureGroup | null>(null)

  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasPolygon, setHasPolygon] = useState(false)

  useEffect(() => {
    if (!mapRef.current || mapInstance.current) return

    const map = L.map(mapRef.current).setView(
      [warehouse.lat, warehouse.lng],
      13
    )
    mapInstance.current = map

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap contributors',
      maxZoom: 19,
    }).addTo(map)

    L.marker([warehouse.lat, warehouse.lng])
      .addTo(map)
      .bindPopup(warehouse.name)

    const items = new L.FeatureGroup()
    drawnItems.current = items
    map.addLayer(items)

    // If the warehouse already has a saved service area, load it so the
    // admin edits the existing shape instead of starting from scratch.
    if (warehouse.service_area) {
      try {
        const geo = JSON.parse(warehouse.service_area)
        const layer = L.geoJSON(geo)
        layer.eachLayer((l) => items.addLayer(l))
        setHasPolygon(true)
        const bounds = items.getBounds()
        if (bounds.isValid()) map.fitBounds(bounds, { padding: [40, 40] })
      } catch {
        // ignore malformed existing geometry, admin can just draw a fresh one
      }
    }

    const drawControl = new (L as any).Control.Draw({
      position: 'topright',
      draw: {
        polygon: {
          allowIntersection: false,
          showArea: true,
          shapeOptions: { color: '#6366f1' },
        },
        polyline: false,
        rectangle: false,
        circle: false,
        circlemarker: false,
        marker: false,
      },
      edit: {
        featureGroup: items,
        remove: true,
      },
    })
    map.addControl(drawControl)

    map.on((L as any).Draw.Event.CREATED, (e: any) => {
      // Only one service-area polygon per warehouse - a new shape replaces
      // any previous one instead of stacking up.
      items.clearLayers()
      items.addLayer(e.layer)
      setHasPolygon(true)
    })

    map.on((L as any).Draw.Event.EDITED, () => {
      setHasPolygon(items.getLayers().length > 0)
    })

    map.on((L as any).Draw.Event.DELETED, () => {
      setHasPolygon(items.getLayers().length > 0)
    })

    return () => {
      map.remove()
      mapInstance.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function handleSave() {
    if (!drawnItems.current || drawnItems.current.getLayers().length === 0) {
      setError('Draw a polygon on the map first, then save.')
      return
    }
    const layer = drawnItems.current.getLayers()[0] as L.Polygon
    const geojson = JSON.stringify((layer.toGeoJSON() as any).geometry)

    setIsSaving(true)
    setError(null)
    try {
      await setWarehouseServiceArea(warehouse.id, geojson)
      onSaved()
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to save service area.')
    } finally {
      setIsSaving(false)
    }
  }

  function handleClear() {
    drawnItems.current?.clearLayers()
    setHasPolygon(false)
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-4xl overflow-hidden">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800">
          <div>
            <h2 className="text-base font-semibold">
              Service area — {warehouse.name}
            </h2>
            <p className="text-xs text-slate-400 mt-0.5">
              Use the polygon tool (top-right of the map) to draw the delivery
              boundary. Draw a new shape to replace the existing one.
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-100 text-lg leading-none"
          >
            ×
          </button>
        </div>

        <div
          ref={mapRef}
          className="w-full"
          style={{ height: '60vh', background: '#0f172a' }}
        />

        <div className="flex items-center justify-between px-6 py-4 border-t border-slate-800">
          <div className="text-xs text-slate-400">
            {error ? (
              <span className="text-red-400">{error}</span>
            ) : hasPolygon ? (
              'Polygon ready to save.'
            ) : (
              'No polygon drawn yet.'
            )}
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={handleClear}
              disabled={!hasPolygon}
              className="px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-slate-100 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Clear
            </button>
            <button
              onClick={handleSave}
              disabled={isSaving || !hasPolygon}
              className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSaving ? 'Saving...' : 'Save service area'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
