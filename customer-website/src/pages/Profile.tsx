import { useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { updateProfile } from '../api/auth'

export default function Profile() {
  const { user, setUser } = useAuth()
  const [name, setName] = useState(user?.name ?? '')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [savedMsg, setSavedMsg] = useState(false)

  if (!user) {
    return <div className="max-w-md mx-auto px-6 py-16 text-ink/50">Loading...</div>
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSavedMsg(false)
    if (!name.trim()) {
      setError('Name cannot be empty.')
      return
    }
    setIsSubmitting(true)
    try {
      const updated = await updateProfile(name.trim())
      setUser(updated)
      setSavedMsg(true)
      setTimeout(() => setSavedMsg(false), 2500)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to update profile.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-md mx-auto px-6 py-10">
      <h1 className="font-display text-3xl font-600 mb-6">Your profile</h1>

      <form onSubmit={handleSubmit} className="border border-line rounded-xl p-5 space-y-4">
        <div>
          <label className="block text-xs font-mono uppercase tracking-widest text-ink/50 mb-1">
            Phone
          </label>
          <p className="text-sm text-ink/60 px-3 py-2 border border-line rounded-lg bg-line/10">
            {user.phone}
          </p>
          <p className="text-xs text-ink/40 mt-1">Phone number can't be changed here.</p>
        </div>

        <div>
          <label className="block text-xs font-mono uppercase tracking-widest text-ink/50 mb-1">
            Name
          </label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
            className="w-full border border-line rounded-lg px-3 py-2 outline-none focus:border-ink"
            required
          />
        </div>

        {error && <p className="text-clay text-sm">{error}</p>}
        {savedMsg && <p className="text-leaf text-sm">Profile updated ✓</p>}

        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-ink text-paper text-sm font-medium px-4 py-2 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
        >
          {isSubmitting ? 'Saving...' : 'Save changes'}
        </button>
      </form>
    </div>
  )
}
