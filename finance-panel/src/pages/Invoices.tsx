import { useEffect, useState } from 'react'
import { searchInvoices, downloadInvoicePDF } from '../api/invoices'
import type { Invoice } from '../types/invoices'

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

export default function Invoices() {
  const [invoiceNumber, setInvoiceNumber] = useState('')
  const [orderId, setOrderId] = useState('')
  const [paymentStatus, setPaymentStatus] = useState('')
  const [page, setPage] = useState(1)
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [downloadingId, setDownloadingId] = useState<number | null>(null)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    searchInvoices({
      invoice_number: invoiceNumber || undefined,
      order_id: orderId || undefined,
      payment_status: paymentStatus || undefined,
      page,
      limit: 20,
    })
      .then((res) => {
        if (!cancelled) {
          setInvoices(res.invoices)
          setTotalPages(res.total_pages)
          setTotal(res.total)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.response?.data?.error ?? 'Could not load invoices.')
          setInvoices([])
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [invoiceNumber, orderId, paymentStatus, page])

  async function handleDownload(invoice: Invoice) {
    setDownloadingId(invoice.id)
    try {
      await downloadInvoicePDF(invoice.id, invoice.invoice_number)
    } catch {
      setError('Could not download invoice PDF.')
    } finally {
      setDownloadingId(null)
    }
  }

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-lg font-semibold">Invoices</h1>
        <p className="text-sm text-slate-500">Search and download tax invoices.</p>
      </div>

      <div className="flex flex-wrap items-center gap-2 mb-6 text-sm">
        <input
          type="text"
          placeholder="Invoice number"
          value={invoiceNumber}
          onChange={(e) => { setPage(1); setInvoiceNumber(e.target.value) }}
          className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5"
        />
        <input
          type="text"
          placeholder="Order ID"
          value={orderId}
          onChange={(e) => { setPage(1); setOrderId(e.target.value) }}
          className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 w-28"
        />
        <select
          value={paymentStatus}
          onChange={(e) => { setPage(1); setPaymentStatus(e.target.value) }}
          className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5"
        >
          <option value="">All payment statuses</option>
          <option value="paid">Paid</option>
          <option value="pending">Pending</option>
          <option value="failed">Failed</option>
        </select>
      </div>

      {isLoading && <p className="text-sm text-slate-500">Loading invoices...</p>}
      {!isLoading && error && (
        <div className="border border-slate-800 rounded-xl p-6 text-sm text-slate-500">{error}</div>
      )}

      {!isLoading && !error && (
        <>
          <div className="border border-slate-800 rounded-xl overflow-hidden mb-4">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-2 font-medium">Invoice #</th>
                  <th className="px-4 py-2 font-medium">Order</th>
                  <th className="px-4 py-2 font-medium">Customer</th>
                  <th className="px-4 py-2 font-medium">Date</th>
                  <th className="px-4 py-2 font-medium text-right">Total</th>
                  <th className="px-4 py-2 font-medium text-right">PDF</th>
                </tr>
              </thead>
              <tbody>
                {invoices.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-4 py-6 text-center text-slate-500">
                      No invoices found.
                    </td>
                  </tr>
                )}
                {invoices.map((inv) => (
                  <tr key={inv.id} className="border-t border-slate-800">
                    <td className="px-4 py-2">{inv.invoice_number}</td>
                    <td className="px-4 py-2">#{inv.order_id}</td>
                    <td className="px-4 py-2">{inv.customer_name}</td>
                    <td className="px-4 py-2">{new Date(inv.generated_at).toLocaleDateString('en-IN')}</td>
                    <td className="px-4 py-2 text-right">{formatCurrency(inv.total_amount)}</td>
                    <td className="px-4 py-2 text-right">
                      <button
                        onClick={() => handleDownload(inv)}
                        disabled={downloadingId === inv.id}
                        className="text-emerald-400 hover:text-emerald-300 text-xs disabled:opacity-50"
                      >
                        {downloadingId === inv.id ? 'Downloading...' : 'Download'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between text-sm text-slate-500">
            <span>{total} total invoices</span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-2 py-1 border border-slate-700 rounded disabled:opacity-40"
              >
                Prev
              </button>
              <span>Page {page} of {totalPages || 1}</span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="px-2 py-1 border border-slate-700 rounded disabled:opacity-40"
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
