import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import apiClient from '../api/client'

interface AnalyticsSummary {
  total_users: number
  total_orders: number
  total_sales: number
  pending_orders: number
  delivered_orders: number
  cancelled_orders: number
}

export default function Dashboard() {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function loadSummary() {
      try {
        const { data } = await apiClient.get<{ summary: AnalyticsSummary }>(
          '/admin/analytics/summary'
        )
        if (!cancelled) setSummary(data.summary)
      } catch (err: any) {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Failed to load analytics.')
        }
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    loadSummary()
    return () => {
      cancelled = true
    }
  }, [])

  const cards = summary
    ? [
        { label: 'Total Users', value: summary.total_users },
        { label: 'Total Orders', value: summary.total_orders },
        { label: 'Total Sales', value: `₹${summary.total_sales}` },
        { label: 'Pending Orders', value: summary.pending_orders },
        { label: 'Delivered Orders', value: summary.delivered_orders },
        { label: 'Cancelled Orders', value: summary.cancelled_orders },
      ]
    : []

  return (
    <Layout>
      <div className="p-8">
        <h1 className="text-xl font-semibold mb-6">Overview</h1>

        {isLoading && <p className="text-slate-400">Loading analytics...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {summary && (
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {cards.map((card) => (
              <div
                key={card.label}
                className="bg-slate-900 rounded-xl p-5 border border-slate-800"
              >
                <p className="text-sm text-slate-400 mb-1">{card.label}</p>
                <p className="text-2xl font-bold">{card.value}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </Layout>
  )
}
