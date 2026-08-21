import { useEffect, useState } from 'react'
import { getFinanceDashboard } from '../api/finance'
import type { FinanceDashboard as FinanceDashboardData } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

export default function FinanceDashboard() {
  const [data, setData] = useState<FinanceDashboardData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getFinanceDashboard()
      .then((res) => {
        if (!cancelled) setData(res)
      })
      .catch((err) => {
        if (!cancelled) setError(err.response?.data?.error ?? 'Could not load dashboard.')
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const totalPendingApprovals = data
    ? data.pending_approvals.expenses + data.pending_approvals.journal_entries + data.pending_approvals.bank_changes
    : 0

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-lg font-semibold">Finance Dashboard</h1>
        <p className="text-sm text-slate-500">
          Unified view of revenue, payables, receivables, GST, and bank balance for the current month.
        </p>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading dashboard...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && data && (
        <>
          {totalPendingApprovals > 0 && (
            <div className="border border-amber-800/40 bg-amber-900/10 rounded-xl px-4 py-3 mb-6 text-sm text-amber-400 flex items-center gap-2">
              <span className="font-medium">{totalPendingApprovals}</span>
              <span>
                item{totalPendingApprovals !== 1 ? 's' : ''} waiting on maker-checker approval
                {' '}({data.pending_approvals.expenses} expense{data.pending_approvals.expenses !== 1 ? 's' : ''},{' '}
                {data.pending_approvals.journal_entries} journal entr{data.pending_approvals.journal_entries !== 1 ? 'ies' : 'y'},{' '}
                {data.pending_approvals.bank_changes} bank change{data.pending_approvals.bank_changes !== 1 ? 's' : ''})
              </span>
            </div>
          )}

          <h2 className="text-sm font-semibold mb-3 text-slate-300">This Month</h2>
          <div className="grid grid-cols-3 gap-4 mb-8">
            <Card label="Total Revenue" value={formatCurrency(data.revenue.total_revenue)} />
            <Card label="COGS" value={formatCurrency(data.revenue.cogs)} />
            <Card label="Gross Profit" value={formatCurrency(data.revenue.gross_profit)} tone="emerald" />
            <Card label="Expenses" value={formatCurrency(data.revenue.expenses)} />
            <Card label="Net Profit" value={formatCurrency(data.revenue.net_profit)} tone={data.revenue.net_profit >= 0 ? 'emerald' : 'red'} />
            <Card label="Bank Balance" value={formatCurrency(data.bank_balance)} tone={data.bank_balance >= 0 ? 'emerald' : 'red'} />
          </div>

          <div className="grid grid-cols-2 gap-6 mb-8">
            <div>
              <h2 className="text-sm font-semibold mb-3 text-slate-300">Payables & Receivables</h2>
              <div className="grid grid-cols-1 gap-4">
                <Card label="Vendor Payable (AP)" value={formatCurrency(data.accounts_payable.vendor_payable)} tone="amber" />
                <Card label="Gateway Pending (AR)" value={formatCurrency(data.accounts_receivable.gateway_pending)} tone="amber" />
                <Card label="COD Pending (AR)" value={formatCurrency(data.accounts_receivable.cod_pending)} tone="amber" />
              </div>
            </div>
            <div>
              <h2 className="text-sm font-semibold mb-3 text-slate-300">GST</h2>
              <div className="grid grid-cols-1 gap-4">
                <Card label="Output GST (Sales)" value={formatCurrency(data.gst.output_gst)} />
                <Card label="Vendor GST (ITC)" value={formatCurrency(data.gst.vendor_gst)} />
                <Card label="Net GST Payable" value={formatCurrency(data.gst.net_gst_payable)} tone={data.gst.net_gst_payable >= 0 ? 'amber' : 'emerald'} />
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

function Card({ label, value, tone }: { label: string; value: string; tone?: 'emerald' | 'amber' | 'red' }) {
  const toneClass = tone === 'emerald' ? 'text-emerald-400' : tone === 'amber' ? 'text-amber-400' : tone === 'red' ? 'text-red-400' : ''
  return (
    <div className="border border-slate-800 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className={`text-lg font-semibold ${toneClass}`}>{value}</p>
    </div>
  )
}
