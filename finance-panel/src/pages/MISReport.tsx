import { useEffect, useState } from 'react'
import { getWeeklyMIS, updateMISExpenseApproval, exportWeeklyMIS, exportMonthlyMIS } from '../api/finance'
import type { WeeklyMIS, MISExpenseApproval } from '../types/finance'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

function formatPct(value: number) {
  const sign = value > 0 ? '+' : ''
  return `${sign}${value.toFixed(1)}%`
}

function mondayISO() {
  const d = new Date()
  const day = d.getDay()
  const diff = day === 0 ? -6 : 1 - day
  d.setDate(d.getDate() + diff)
  return d.toISOString().slice(0, 10)
}

export default function MISReport() {
  const [weekStart, setWeekStart] = useState(mondayISO())
  const [data, setData] = useState<WeeklyMIS | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    getWeeklyMIS(weekStart)
      .then((res) => {
        if (!cancelled) setData(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load MIS report.')
          setData(null)
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }

  useEffect(load, [weekStart])

  const [isExporting, setIsExporting] = useState(false)
  const [isExportingMonthly, setIsExportingMonthly] = useState(false)
  const [monthValue, setMonthValue] = useState(new Date().toISOString().slice(0, 7))
  const handleExportMonthly = async () => {
    setIsExportingMonthly(true)
    try {
      const blob = await exportMonthlyMIS(monthValue)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `monthly-mis-${monthValue}.xlsx`
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
    } finally {
      setIsExportingMonthly(false)
    }
  }
  const handleExport = async () => {
    setIsExporting(true)
    try {
      const blob = await exportWeeklyMIS(weekStart)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `weekly-mis-${weekStart}.xlsx`
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Weekly MIS Report</h1>
          <p className="text-sm text-slate-500">
            {data ? `${data.week_start} to ${data.week_end}` : 'Revenue, vendor expenses, and settlement summary.'}
          </p>
        </div>

        <div className="flex items-center gap-2 text-sm">
          <label className="text-slate-500">Week starting</label>
          <input
            type="date"
            value={weekStart}
            onChange={(e) => setWeekStart(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5"
          />
          <button
            onClick={handleExport}
            disabled={isExporting}
            className="bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg px-3 py-1.5 text-sm"
          >
            {isExporting ? 'Exporting...' : 'Export Excel'}
          </button>
          <input
            type="month"
            value={monthValue}
            onChange={(e) => setMonthValue(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5"
          />
          <button
            onClick={handleExportMonthly}
            disabled={isExportingMonthly}
            className="bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg px-3 py-1.5 text-sm"
          >
            {isExportingMonthly ? 'Exporting...' : 'Export Monthly Excel'}
          </button>
        </div>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading MIS report...</p>}

      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && data && (
        <>
          <MISTable title="Revenue MIS" rows={data.revenue_mis} />
          <MISTable title="Vendor Expense MIS" rows={data.vendor_expense_mis} />

          <h2 className="text-sm font-semibold mb-3">Vendor Settlement</h2>
          <div className="border border-slate-800 rounded-xl overflow-hidden mb-8">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Vendor</th>
                  <th className="px-4 py-2 font-medium text-right">Gross Sales</th>
                  <th className="px-4 py-2 font-medium text-right">Other Charges</th>
                  <th className="px-4 py-2 font-medium text-right">Net Payable</th>
                  <th className="px-4 py-2 font-medium text-right">Paid</th>
                  <th className="px-4 py-2 font-medium text-right">Balance</th>
                  <th className="px-4 py-2 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {data.vendor_settlement.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-4 py-6 text-center text-slate-500">
                      No vendor settlement activity for this week.
                    </td>
                  </tr>
                )}
                {data.vendor_settlement.map((row) => (
                  <tr key={row.vendor_id} className="border-t border-slate-800">
                    <td className="px-4 py-2">{row.vendor_name}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.gross_sales)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.other_charges)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.net_payable)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.amount_paid)}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(row.balance)}</td>
                    <td className="px-4 py-2">
                      <span
                        className={
                          row.status === 'Paid'
                            ? 'text-emerald-400'
                            : row.status === 'Partial'
                              ? 'text-amber-400'
                              : 'text-slate-400'
                        }
                      >
                        {row.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <ExpenseApprovalTable rows={data.expense_approval} onSaved={load} />
        </>
      )}
    </div>
  )
}

function MISTable({ title, rows }: { title: string; rows: { label: string; current: number; previous: number; growth_pct: number }[] }) {
  return (
    <>
      <h2 className="text-sm font-semibold mb-3">{title}</h2>
      <div className="border border-slate-800 rounded-xl overflow-hidden mb-8">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-slate-900 text-slate-400 text-left">
              <th className="px-4 py-2 font-medium">Line Item</th>
              <th className="px-4 py-2 font-medium text-right">This Week</th>
              <th className="px-4 py-2 font-medium text-right">Last Week</th>
              <th className="px-4 py-2 font-medium text-right">Growth</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              const isPct = row.label.includes('%')
              return (
                <tr key={row.label} className="border-t border-slate-800">
                  <td className="px-4 py-2">{row.label}</td>
                  <td className="px-4 py-2 text-right">
                    {isPct ? `${row.current.toFixed(1)}%` : formatCurrency(row.current)}
                  </td>
                  <td className="px-4 py-2 text-right text-slate-500">
                    {isPct ? `${row.previous.toFixed(1)}%` : formatCurrency(row.previous)}
                  </td>
                  <td
                    className={`px-4 py-2 text-right ${
                      row.growth_pct > 0 ? 'text-emerald-400' : row.growth_pct < 0 ? 'text-red-400' : 'text-slate-500'
                    }`}
                  >
                    {formatPct(row.growth_pct)}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </>
  )
}

function ExpenseApprovalTable({ rows, onSaved }: { rows: MISExpenseApproval[]; onSaved: () => void }) {
  const [editingId, setEditingId] = useState<number | null>(null)
  const [draft, setDraft] = useState<MISExpenseApproval | null>(null)
  const [isSaving, setIsSaving] = useState(false)

  const startEdit = (row: MISExpenseApproval) => {
    setEditingId(row.id)
    setDraft({ ...row })
  }

  const cancelEdit = () => {
    setEditingId(null)
    setDraft(null)
  }

  const save = async () => {
    if (!draft) return
    setIsSaving(true)
    try {
      await updateMISExpenseApproval(draft.id, {
        category: draft.category,
        up_to_25k: draft.up_to_25k,
        range_25k_1l: draft.range_25k_1l,
        range_1l_5l: draft.range_1l_5l,
        above_5l: draft.above_5l,
        required_documents: draft.required_documents,
        approver: draft.approver,
      })
      setEditingId(null)
      setDraft(null)
      onSaved()
    } finally {
      setIsSaving(false)
    }
  }

  const field = (key: keyof MISExpenseApproval) => (
    <input
      value={(draft?.[key] as string) ?? ''}
      onChange={(e) => setDraft((d) => (d ? { ...d, [key]: e.target.value } : d))}
      className="bg-slate-800 border border-slate-700 rounded px-2 py-1 w-full text-xs"
    />
  )

  return (
    <>
      <h2 className="text-sm font-semibold mb-3">Expense Approval Matrix</h2>
      <div className="border border-slate-800 rounded-xl overflow-hidden mb-8">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-slate-900 text-slate-400 text-left">
              <th className="px-4 py-2 font-medium">Category</th>
              <th className="px-4 py-2 font-medium">Up to 25k</th>
              <th className="px-4 py-2 font-medium">25k - 1L</th>
              <th className="px-4 py-2 font-medium">1L - 5L</th>
              <th className="px-4 py-2 font-medium">Above 5L</th>
              <th className="px-4 py-2 font-medium">Required Docs</th>
              <th className="px-4 py-2 font-medium">Approver</th>
              <th className="px-4 py-2 font-medium"></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              const isEditing = editingId === row.id
              return (
                <tr key={row.id} className="border-t border-slate-800 align-top">
                  <td className="px-4 py-2">{row.category}</td>
                  {isEditing ? (
                    <>
                      <td className="px-4 py-2">{field('up_to_25k')}</td>
                      <td className="px-4 py-2">{field('range_25k_1l')}</td>
                      <td className="px-4 py-2">{field('range_1l_5l')}</td>
                      <td className="px-4 py-2">{field('above_5l')}</td>
                      <td className="px-4 py-2">{field('required_documents')}</td>
                      <td className="px-4 py-2">{field('approver')}</td>
                      <td className="px-4 py-2 whitespace-nowrap">
                        <button
                          onClick={save}
                          disabled={isSaving}
                          className="text-emerald-400 hover:text-emerald-300 mr-2 text-xs"
                        >
                          {isSaving ? 'Saving...' : 'Save'}
                        </button>
                        <button onClick={cancelEdit} className="text-slate-500 hover:text-slate-400 text-xs">
                          Cancel
                        </button>
                      </td>
                    </>
                  ) : (
                    <>
                      <td className="px-4 py-2 text-slate-400">{row.up_to_25k || '-'}</td>
                      <td className="px-4 py-2 text-slate-400">{row.range_25k_1l || '-'}</td>
                      <td className="px-4 py-2 text-slate-400">{row.range_1l_5l || '-'}</td>
                      <td className="px-4 py-2 text-slate-400">{row.above_5l || '-'}</td>
                      <td className="px-4 py-2 text-slate-400">{row.required_documents || '-'}</td>
                      <td className="px-4 py-2 text-slate-400">{row.approver || '-'}</td>
                      <td className="px-4 py-2">
                        <button onClick={() => startEdit(row)} className="text-sky-400 hover:text-sky-300 text-xs">
                          Edit
                        </button>
                      </td>
                    </>
                  )}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </>
  )
}
