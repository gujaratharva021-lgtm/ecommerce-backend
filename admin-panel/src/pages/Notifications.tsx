import { useState } from 'react'
import Layout from '../components/Layout'
import { broadcastNotification } from '../api/admin'

export default function Notifications() {
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [isSending, setIsSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)

  async function handleSend(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSuccessMsg(null)

    if (!title.trim() || !body.trim()) {
      setError('Title and message are required.')
      return
    }

    const confirmed = confirm(
      'Send this notification to ALL users with the app installed? This cannot be undone.'
    )
    if (!confirmed) return

    setIsSending(true)
    try {
      await broadcastNotification(title.trim(), body.trim())
      setSuccessMsg('Notification broadcast queued successfully.')
      setTitle('')
      setBody('')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to send notification.')
    } finally {
      setIsSending(false)
    }
  }

  return (
    <Layout>
      <div className="p-8 max-w-2xl">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Notifications</h1>
          <p className="text-sm text-slate-400 mt-1">
            Broadcast a push notification to every registered device
          </p>
        </div>

        <form
          onSubmit={handleSend}
          className="border border-slate-800 rounded-xl p-6 bg-slate-900 space-y-4"
        >
          <div>
            <label className="block text-sm text-slate-400 mb-1">Title</label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Flash Sale Today!"
              maxLength={100}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
            />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Message</label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="e.g. Get 20% off on all electronics, today only."
              maxLength={300}
              rows={4}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500 resize-none"
            />
            <p className="text-xs text-slate-500 mt-1">{body.length}/300</p>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}
          {successMsg && <p className="text-emerald-400 text-sm">{successMsg}</p>}

          <button
            type="submit"
            disabled={isSending}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium disabled:opacity-40 transition-colors"
          >
            {isSending ? 'Sending...' : 'Send to all users'}
          </button>
        </form>

        <p className="text-xs text-slate-500 mt-4">
          Sent notifications are recorded in{' '}
          <a href="/audit-logs" className="text-indigo-400 hover:text-indigo-300">
            Audit Logs
          </a>{' '}
          under the "broadcast_notification" action.
        </p>
      </div>
    </Layout>
  )
}
