import { useState } from 'react'
import Layout from '../components/Layout'
import { creditWallet } from '../api/admin'

export default function WalletCredit() {
  const [userId, setUserId] = useState('')
  const [amount, setAmount] = useState('')
  const [note, setNote] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<{ user_id: number; balance: number } | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setResult(null)

    const uid = parseInt(userId, 10)
    const amt = parseFloat(amount)

    if (!uid || uid <= 0) {
      setError('Please enter a valid user ID.')
      return
    }
    if (!amt || amt <= 0) {
      setError('Amount must be greater than 0.')
      return
    }

    setIsSaving(true)
    try {
      const res = await creditWallet(uid, amt, note.trim() || undefined)
      setResult({ user_id: res.wallet?.user_id ?? uid, balance: res.wallet?.balance ?? 0 })
      setUserId('')
      setAmount('')
      setNote('')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to credit wallet.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Wallet Credit</h1>
          <p className="text-sm text-slate-400 mt-1">
            Manually credit a customer's wallet (cashback, goodwill, or support correction)
          </p>
        </div>

        <div className="max-w-md border border-slate-800 rounded-xl p-6 bg-slate-900">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="text-xs text-slate-400 block mb-1">User ID</label>
              <input
                type="number"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                placeholder="e.g. 1"
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Amount (?)</label>
              <input
                type="number"
                step="0.01"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="e.g. 100"
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Note (optional)</label>
              <input
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="e.g. Goodwill credit for delayed order"
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>

            {error && <p className="text-red-400 text-xs">{error}</p>}
            {result && (
              <p className="text-emerald-400 text-xs">
                Credited successfully. User #{result.user_id} new balance: ?{result.balance}
              </p>
            )}

            <button
              type="submit"
              disabled={isSaving}
              className="w-full py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
            >
              {isSaving ? 'Crediting...' : 'Credit Wallet'}
            </button>
          </form>
        </div>
      </div>
    </Layout>
  )
}
