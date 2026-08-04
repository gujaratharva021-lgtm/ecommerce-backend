import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { getAnalyticsSummary, getProductPerformance } from '../api/admin'
import type { AnalyticsSummary, ProductPerformance } from '../types/admin'

export default function Analytics() {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [performance, setPerformance] = useState<ProductPerformance[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const [summaryRes, perfRes] = await Promise.all([
        getAnalyticsSummary(),
        getProductPerformance(),
      ])
      setSummary(summaryRes.summary ?? null)
      setPerformance(perfRes.products ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load analytics.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const cards = [
    { label: 'Total Sales', value: summary?.total_sales, prefix: '₹' },
    { label: 'Total Orders', value: summary?.total_orders },
    { label: 'Total Users', value: summary?.total_users },
    { label: 'Pending Orders', value: summary?.pending_orders },
    { label: 'Delivered Orders', value: summary?.delivered_orders },
    { label: 'Cancelled Orders', value: summary?.cancelled_orders },
  ]

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Analytics</h1>
          <p className="text-sm text-slate-400 mt-1">Store performance overview</p>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && (
          <>
            <div className="grid grid-cols-3 gap-4 mb-8">
              {cards.map((c) => (
                <div
                  key={c.label}
                  className="border border-slate-800 rounded-xl p-5 bg-slate-900"
                >
                  <p className="text-xs text-slate-400 mb-2">{c.label}</p>
                  <p className="text-2xl font-semibold">
                    {c.prefix ?? ''}
                    {c.value ?? 0}
                  </p>
                </div>
              ))}
            </div>

            <h2 className="text-sm font-medium text-slate-300 mb-3">
              Top Product Performance
            </h2>

            {performance.length === 0 ? (
              <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
                No product performance data yet.
              </div>
            ) : (
              <div className="border border-slate-800 rounded-xl overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-slate-900 text-slate-400 text-left">
                      <th className="px-4 py-3 font-medium">Product</th>
                      <th className="px-4 py-3 font-medium">Units Sold</th>
                      <th className="px-4 py-3 font-medium">Revenue</th>
                    </tr>
                  </thead>
                  <tbody>
                    {performance.map((p) => (
                      <tr key={p.product_id} className="border-t border-slate-800">
                        <td className="px-4 py-3">{p.product_name}</td>
                        <td className="px-4 py-3">{p.units_sold}</td>
                        <td className="px-4 py-3">₹{p.total_revenue}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </>
        )}
      </div>
    </Layout>
  )
}
