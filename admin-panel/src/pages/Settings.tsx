import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { getSettings, updateSettings } from '../api/admin'
import { SETTINGS_LABELS } from '../types/admin'

export default function Settings() {
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [original, setOriginal] = useState<Record<string, string>>({})
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await getSettings()
      setSettings(res.settings ?? {})
      setOriginal(res.settings ?? {})
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load settings.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  function handleChange(key: string, value: string) {
    setSettings((prev) => ({ ...prev, [key]: value }))
    setSuccessMsg(null)
  }

  const changedKeys = Object.keys(settings).filter((k) => settings[k] !== original[k])
  const hasChanges = changedKeys.length > 0

  async function handleSave() {
    setIsSaving(true)
    setError(null)
    setSuccessMsg(null)
    try {
      const changedSettings: Record<string, string> = {}
      changedKeys.forEach((k) => {
        changedSettings[k] = settings[k]
      })
      await updateSettings(changedSettings)
      setOriginal(settings)
      setSuccessMsg('Settings updated successfully.')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to update settings. You may not have permission for this action.')
    } finally {
      setIsSaving(false)
    }
  }

  function handleReset() {
    setSettings(original)
    setSuccessMsg(null)
  }

  const keys = Object.keys(settings)

  return (
    <Layout>
      <div className="p-8 max-w-2xl">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Settings</h1>
          <p className="text-sm text-slate-400 mt-1">
            Business configuration &middot; changes require Settings Management permission
          </p>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400 mb-4">{error}</p>}
        {successMsg && <p className="text-emerald-400 mb-4">{successMsg}</p>}

        {!isLoading && keys.length > 0 && (
          <div className="border border-slate-800 rounded-xl p-6 bg-slate-900 space-y-4">
            {keys.map((key) => (
              <div key={key}>
                <label className="block text-sm text-slate-400 mb-1">
                  {SETTINGS_LABELS[key] ?? key}
                </label>
                <input
                  type="text"
                  value={settings[key]}
                  onChange={(e) => handleChange(key, e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                />
              </div>
            ))}

            <div className="flex items-center gap-3 pt-4 border-t border-slate-800">
              <button
                disabled={!hasChanges || isSaving}
                onClick={handleSave}
                className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-sm font-medium disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {isSaving ? 'Saving...' : 'Save Changes'}
              </button>
              <button
                disabled={!hasChanges || isSaving}
                onClick={handleReset}
                className="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-sm disabled:opacity-40"
              >
                Reset
              </button>
              {hasChanges && (
                <span className="text-xs text-slate-500">
                  {changedKeys.length} unsaved change{changedKeys.length !== 1 ? 's' : ''}
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </Layout>
  )
}
