import { useEffect, useRef, useState } from 'react'
import { Html5Qrcode } from 'html5-qrcode'

interface CameraScannerProps {
  onDetected: (code: string) => void
  onClose: () => void
}

// Full-screen camera scanner modal. Opens the device's rear camera where
// available, decodes barcodes/QR codes live, and calls onDetected once with
// the first successful read (the caller is responsible for closing /
// re-opening for the next scan, so a stray extra frame can't fire twice).
export default function CameraScanner({ onDetected, onClose }: CameraScannerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const hasFiredRef = useRef(false)
  const [error, setError] = useState<string | null>(null)
  const [isStarting, setIsStarting] = useState(true)

  useEffect(() => {
    const containerId = 'camera-scanner-viewport'
    const scanner = new Html5Qrcode(containerId, { verbose: false })
    scannerRef.current = scanner

    scanner
      .start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: { width: 250, height: 150 } },
        (decodedText) => {
          if (hasFiredRef.current) return
          hasFiredRef.current = true
          onDetected(decodedText)
        },
        () => {
          // Per-frame "nothing decoded yet" callback - expected constantly
          // while the camera is pointed at anything that isn't a barcode,
          // so it's intentionally not surfaced as an error.
        }
      )
      .then(() => setIsStarting(false))
      .catch((err) => {
        setIsStarting(false)
        setError(
          err?.message?.includes('Permission')
            ? 'Camera permission was denied. Allow camera access to scan, or type the barcode manually.'
            : 'Could not start the camera. Your device may not support camera scanning here.'
        )
      })

    return () => {
      scanner
        .stop()
        .then(() => scanner.clear())
        .catch(() => {
          /* already stopped / never started - fine to ignore on unmount */
        })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 w-full max-w-sm">
        <div className="flex items-center justify-between mb-3">
          <p className="text-sm font-medium">Scan Barcode</p>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-200 text-sm">
            Close
          </button>
        </div>

        {error ? (
          <div className="text-xs text-rose-400 border border-rose-900 bg-rose-950/40 rounded-lg px-3 py-3">
            {error}
          </div>
        ) : (
          <>
            {isStarting && <p className="text-xs text-slate-400 mb-2">Starting camera...</p>}
            <div id="camera-scanner-viewport" ref={containerRef} className="rounded-lg overflow-hidden bg-black" />
            <p className="text-xs text-slate-500 mt-3 text-center">
              Point the camera at the product's barcode. It scans automatically.
            </p>
          </>
        )}
      </div>
    </div>
  )
}
