import { useEffect, useState } from 'react'
import { getSettings, updateSettings } from '../api/settings'
import type { Settings } from '../types/settings'

const FIELD_LABELS: Record<keyof Settings, string> = {
  free_delivery_threshold: 'Free Delivery Threshold (Rs.)',
  flat_delivery_charge: 'Flat Delivery Charge (Rs.)',
  platform_fee: 'Platform Fee (Rs.)',
  min_order_amount: 'Minimum Order Amount (Rs.)',
  cancellation_window_minutes: 'Cancellation Window (minutes)',
  company_name: 'Company Name',
  support_phone: 'Support Phone',
  support_email: 'Support Email',
  gst_percentage: 'GST Percentage (%)',
}

export default function SettingsPage() {
  const [values, setValues] = useState<Settings | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [savedMessage, setSavedMessage] = useState<string | null>(null)

  useEffect(() => {
    getSettings()
      .then(setValues)
      .catch((err) => setError(err.response?.data?.error ?? 'Could not load settings.'))
      .finally(() => setIsLoading(false))
  }, [])

  function handleChange(key: keyof Settings, value: string) {
    setValues((prev) => (prev ? { ...prev, [key]: value } : prev))
    setSavedMessage(null)
  }

  async function handleSave() {
    if (!values) return
    setIsSaving(true)
    setError(null)
    setSavedMessage(null)
    try {
      await updateSettings(values)
      setSavedMessage('Settings saved.')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Could not save settings.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="p-8 max-w-xl">
      <div className="mb-6">
        <h1 className="text-lg font-semibold">Settings</h1>
        <p className="text-sm text-slate-500">Store-wide configuration.</p>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading settings...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500 mb-4">{error}</div>
      )}

      {!isLoading && values && (
        <div className="space-y-4">
          {(Object.keys(FIELD_LABELS) as (keyof Settings)[]).map((key) => (
            <div key={key}>
              <label className="text-xs text-slate-400 block mb-1">{FIELD_LABELS[key]}</label>
              <input
                type="text"
                value={values[key] ?? ''}
                onChange={(e) => handleChange(key, e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
          ))}

          {savedMessage && <p className="text-emerald-400 text-xs">{savedMessage}</p>}

          <button
            onClick={handleSave}
            disabled={isSaving}
            className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      )}
    </div>
  )
}
