import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { searchAdminInvoices, getAdminInvoiceById, downloadAdminInvoicePDF } from '../api/admin'
import type { AdminInvoiceListItem, AdminInvoice } from '../types/admin'

function fmtMoney(n: number) {
  return '\u20b9' + Number(n ?? 0).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export default function Invoices() {
  const [rows, setRows] = useState<AdminInvoiceListItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [invoiceNumber, setInvoiceNumber] = useState('')
  const [orderId, setOrderId] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)

  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [detail, setDetail] = useState<AdminInvoice | null>(null)
  const [isLoadingDetail, setIsLoadingDetail] = useState(false)
  const [isDownloading, setIsDownloading] = useState(false)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await searchAdminInvoices({
        invoice_number: invoiceNumber || undefined,
        order_id: orderId || undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
        page,
        limit: 20,
      })
      setRows(res.invoices ?? [])
      setTotalPages(res.total_pages ?? 1)
      setTotal(res.total ?? 0)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load invoices.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page])

  function handleSearchSubmit(e: React.FormEvent) {
    e.preventDefault()
    setPage(1)
    load()
  }

  async function openDetail(id: number) {
    setSelectedId(id)
    setIsLoadingDetail(true)
    setDetail(null)
    try {
      const res = await getAdminInvoiceById(id)
      setDetail(res)
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to load invoice detail.')
    } finally {
      setIsLoadingDetail(false)
    }
  }

  async function handleDownload(id: number, invoiceNum?: string) {
    setIsDownloading(true)
    try {
      await downloadAdminInvoicePDF(id, invoiceNum)
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to download invoice PDF.')
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Invoices</h1>
          <p className="text-sm text-slate-400 mt-1">Search and download tax invoices generated for orders</p>
        </div>

        <form onSubmit={handleSearchSubmit} className="flex flex-wrap items-center gap-2 mb-4">
          <input
            type="text"
            value={invoiceNumber}
            onChange={(e) => setInvoiceNumber(e.target.value)}
            placeholder="Invoice number..."
            className="flex-1 min-w-[180px] bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
          />
          <input
            type="text"
            value={orderId}
            onChange={(e) => setOrderId(e.target.value)}
            placeholder="Order ID..."
            className="w-32 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
          />
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
          />
          <span className="text-slate-500 text-sm">to</span>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
          />
          <button
            type="submit"
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            Search
          </button>
        </form>

        <div className="grid grid-cols-5 gap-6">
          <div className="col-span-3">
            {isLoading && <p className="text-slate-400">Loading...</p>}
            {error && <p className="text-red-400">{error}</p>}

            {!isLoading && !error && rows.length === 0 && (
              <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
                No invoices found.
              </div>
            )}

            {!isLoading && rows.length > 0 && (
              <div className="border border-slate-800 rounded-xl overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-slate-900 text-slate-400 text-xs uppercase">
                    <tr>
                      <th className="text-left px-4 py-3">Invoice</th>
                      <th className="text-left px-4 py-3">Customer</th>
                      <th className="text-right px-4 py-3">Amount</th>
                      <th className="text-left px-4 py-3">Method</th>
                      <th className="text-right px-4 py-3">Download</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    {rows.map((r) => (
                      <tr
                        key={r.id}
                        onClick={() => openDetail(r.id)}
                        className={
                          'cursor-pointer hover:bg-slate-900 transition-colors ' +
                          (selectedId === r.id ? 'bg-slate-900' : '')
                        }
                      >
                        <td className="px-4 py-3">
                          <p className="text-slate-200 font-medium">{r.invoice_number}</p>
                          <p className="text-xs text-slate-500">
                            Order #{r.order_id} &middot; {new Date(r.generated_at).toLocaleString()}
                          </p>
                        </td>
                        <td className="px-4 py-3">
                          <p className="text-slate-300">{r.customer_name || '\u2014'}</p>
                          <p className="text-xs text-slate-500">{r.customer_phone}</p>
                        </td>
                        <td className="px-4 py-3 text-right text-slate-200">{fmtMoney(r.total_amount)}</td>
                        <td className="px-4 py-3">
                          <span className="text-xs uppercase text-slate-400">{r.payment_method}</span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleDownload(r.id, r.invoice_number)
                            }}
                            disabled={isDownloading}
                            className="text-indigo-400 hover:text-indigo-300 text-xs font-medium disabled:opacity-40"
                          >
                            PDF
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {!isLoading && rows.length > 0 && (
              <div className="flex items-center justify-between mt-4 text-sm text-slate-400">
                <span>
                  Page {page} of {totalPages} &middot; {total} total
                </span>
                <div className="flex gap-2">
                  <button
                    disabled={page <= 1}
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    className="px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 disabled:opacity-40"
                  >
                    Prev
                  </button>
                  <button
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                    className="px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 disabled:opacity-40"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </div>

          <div className="col-span-2">
            {!selectedId && (
              <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500 h-full flex items-center justify-center">
                Select an invoice to view details
              </div>
            )}

            {selectedId && (
              <div className="border border-slate-800 rounded-xl overflow-hidden">
                <div className="px-4 py-3 bg-slate-900 border-b border-slate-800 flex items-center justify-between">
                  <p className="text-sm font-medium text-slate-200">
                    {detail?.invoice_number ?? `Invoice #${selectedId}`}
                  </p>
                  <button
                    onClick={() => handleDownload(selectedId, detail?.invoice_number)}
                    disabled={isDownloading}
                    className="px-3 py-1.5 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-xs font-medium disabled:opacity-40 transition-colors"
                  >
                    {isDownloading ? 'Downloading...' : 'Download PDF'}
                  </button>
                </div>

                {isLoadingDetail && <p className="text-slate-400 text-sm p-4">Loading...</p>}

                {!isLoadingDetail && detail && (
                  <div className="p-4 space-y-4">
                    <div className="grid grid-cols-2 gap-3 text-sm">
                      <div>
                        <p className="text-xs text-slate-500">Customer</p>
                        <p className="text-slate-200">{detail.customer_name || '\u2014'}</p>
                        <p className="text-xs text-slate-500">{detail.customer_phone}</p>
                      </div>
                      <div>
                        <p className="text-xs text-slate-500">Order</p>
                        <p className="text-slate-200">#{detail.order_id}</p>
                        <p className="text-xs text-slate-500">{detail.order_status}</p>
                      </div>
                      <div>
                        <p className="text-xs text-slate-500">Delivery Address</p>
                        <p className="text-slate-200">
                          {detail.address_line1}
                          {detail.address_line2 ? `, ${detail.address_line2}` : ''}
                        </p>
                        <p className="text-xs text-slate-500">
                          {detail.address_city}, {detail.address_state} - {detail.address_pincode}
                        </p>
                      </div>
                      <div>
                        <p className="text-xs text-slate-500">Payment</p>
                        <p className="text-slate-200 uppercase">{detail.payment_method}</p>
                        <p className="text-xs text-slate-500">{detail.payment_status}</p>
                      </div>
                    </div>

                    <div className="border-t border-slate-800 pt-4">
                      <p className="text-xs text-slate-500 mb-2">Items</p>
                      <div className="space-y-2">
                        {detail.items?.map((item) => (
                          <div key={item.id} className="flex justify-between text-sm">
                            <span className="text-slate-300">
                              {item.product_name} <span className="text-slate-500">x{item.quantity}</span>
                            </span>
                            <span className="text-slate-200">{fmtMoney(item.price * item.quantity)}</span>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div className="border-t border-slate-800 pt-4 space-y-1.5 text-sm">
                      <div className="flex justify-between">
                        <span className="text-slate-500">Items Amount</span>
                        <span className="text-slate-200">{fmtMoney(detail.items_amount)}</span>
                      </div>
                      {detail.discount_amount > 0 && (
                        <div className="flex justify-between">
                          <span className="text-slate-500">Discount</span>
                          <span className="text-emerald-300">-{fmtMoney(detail.discount_amount)}</span>
                        </div>
                      )}
                      <div className="flex justify-between">
                        <span className="text-slate-500">Delivery Charge</span>
                        <span className="text-slate-200">{fmtMoney(detail.delivery_charge)}</span>
                      </div>
                      {detail.wallet_used > 0 && (
                        <div className="flex justify-between">
                          <span className="text-slate-500">Wallet Used</span>
                          <span className="text-emerald-300">-{fmtMoney(detail.wallet_used)}</span>
                        </div>
                      )}
                      <div className="flex justify-between font-semibold pt-1 border-t border-slate-800">
                        <span>Total</span>
                        <span>{fmtMoney(detail.total_amount)}</span>
                      </div>
                    </div>

                    {detail.seller && (
                      <div className="border-t border-slate-800 pt-4 text-xs text-slate-500">
                        <p className="text-slate-400 mb-1">Seller</p>
                        <p>{detail.seller.company_name || '\u2014'}</p>
                        {detail.seller.gstin && <p>GSTIN: {detail.seller.gstin}</p>}
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </Layout>
  )
}
