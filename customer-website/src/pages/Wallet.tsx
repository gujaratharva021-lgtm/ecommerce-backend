import { useEffect, useState } from 'react'
import { getWallet } from '../api/misc'
import type { WalletResponse } from '../types'

export default function Wallet() {
  const [wallet, setWallet] = useState<WalletResponse | null>(null)

  useEffect(() => {
    getWallet().then(setWallet)
  }, [])

  if (!wallet) {
    return <div className="max-w-2xl mx-auto px-6 py-16 text-ink/50">Loading...</div>
  }

  return (
    <div className="max-w-2xl mx-auto px-6 py-10">
      <h1 className="font-display text-3xl font-600 mb-6">Wallet</h1>

      <div className="border border-line rounded-xl p-6 mb-8 bg-marigold/5">
        <p className="text-sm text-ink/60 mb-1">Available balance</p>
        <p className="font-mono text-4xl font-semibold text-marigold">₹{wallet.balance.toFixed(2)}</p>
      </div>

      <h2 className="font-medium mb-4">Transaction history</h2>
      {wallet.transactions.length === 0 ? (
        <p className="text-ink/50 text-sm">No transactions yet.</p>
      ) : (
        <div className="divide-y divide-line border border-line rounded-xl overflow-hidden">
          {wallet.transactions.map((t) => (
            <div key={t.id} className="flex items-center justify-between p-4">
              <div>
                <p className="text-sm font-medium capitalize">{t.reason.replace(/_/g, ' ')}</p>
                <p className="text-xs text-ink/50">
                  {new Date(t.created_at).toLocaleDateString('en-IN', {
                    day: 'numeric',
                    month: 'short',
                    year: 'numeric',
                  })}
                  {t.note && ` · ${t.note}`}
                </p>
              </div>
              <p className={`font-mono font-semibold ${t.type === 'credit' ? 'text-leaf' : 'text-clay'}`}>
                {t.type === 'credit' ? '+' : '−'}₹{t.amount.toFixed(2)}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
