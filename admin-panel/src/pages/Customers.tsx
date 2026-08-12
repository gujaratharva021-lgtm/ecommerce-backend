import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import Modal from '../components/Modal'
import { listCustomers, getCustomer, blockCustomer, unblockCustomer } from '../api/admin'
import type { CustomerSummary, CustomerDetail } from '../types/admin'

export default function Customers() {
  const [customers, setCustomers] = useState<CustomerSummary[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)

  const [selected, setSelected] = useState<CustomerDetail | null>(null)
  const [isDetailLoading, setIsDetailLoading] = useState(false)
  const [busyId, setBusyId] = useState<number | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const params: Record<string, any> = { page, limit: 20 }
      if (search.trim()) params.search = search.trim()
      if (statusFilter) params.status = statusFilter
      const res = await listCustomers(params)
      setCustomers(res.customers ?? [])
      setTotalPages(res.total_pages ?? 1)
      setTotal(res.total ?? 0)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load customers.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, statusFilter])

  function handleSearchSubmit(e: React.FormEvent) {
    e.preventDefault()
    setPage(1)
    load()
  }

  async function openDetail(id: number) {
    setIsDetailLoading(true)
    setSelected(null)
    try {
      const res = await getCustomer(id)
      setSelected(res)
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to load customer detail.')
    } finally {
      setIsDetailLoading(false)
    }
  }

  async function handleToggleBlock(c: CustomerSummary) {
    setBusyId(c.id)
    try {
      if (c.is_blocked) {
        await unblockCustomer(c.id)
      } else {
        await blockCustomer(c.id)
      }
      setCustomers((prev) =>
        prev.map((u) => (u.id === c.id ? { ...u, is_blocked: !u.is_blocked } : u))
      )
      if (selected && selected.id === c.id) {
        setSelected({ ...selected, is_blocked: !selected.is_blocked })
      }
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update customer status.')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Customers</h1>
            <p className="text-sm text-slate-400 mt-1">{total} customer{total !== 1 ? 's' : ''}</p>
          </div>
        </div>

        <div className="flex items-center gap-3 mb-6">
          <form onSubmit={handleSearchSubmit} className="flex-1 flex gap-2">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search by name or phone..."
              className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
            />
            <button
              type="submit"
              className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-sm font-medium"
            >
              Search
            </button>
          </form>
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value)
              setPage(1)
            }}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
          >
            <option value="">All customers</option>
            <option value="active">Active</option>
            <option value="blocked">Blocked</option>
          </select>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && customers.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No customers found.
          </div>
        )}

        {!isLoading && customers.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Customer</th>
                  <th className="px-4 py-3 font-medium">Phone</th>
                  <th className="px-4 py-3 font-medium">Orders</th>
                  <th className="px-4 py-3 font-medium">Total Spent</th>
                  <th className="px-4 py-3 font-medium">Last Order</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Joined</th>
                  <th className="px-4 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {customers.map((c) => (
                  <tr key={c.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">
                      <button
                        onClick={() => openDetail(c.id)}
                        className="text-indigo-300 hover:underline"
                      >
                        {c.name?.trim() || `Customer #${c.id}`}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-slate-300">{c.phone}</td>
                    <td className="px-4 py-3 text-slate-300">{c.total_orders}</td>
                    <td className="px-4 py-3 text-slate-300">{"\u20B9"}{c.total_spent.toFixed(2)}</td>
                    <td className="px-4 py-3 text-slate-400">
                      {c.last_order_at ? new Date(c.last_order_at).toLocaleDateString() : '\u2014'}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-1 rounded-md text-xs font-medium ${
                          c.is_blocked
                            ? 'bg-red-500/15 text-red-300'
                            : 'bg-emerald-500/15 text-emerald-300'
                        }`}
                      >
                        {c.is_blocked ? 'Blocked' : 'Active'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-400">
                      {new Date(c.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        disabled={busyId === c.id}
                        onClick={() => handleToggleBlock(c)}
                        className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors disabled:opacity-50 ${
                          c.is_blocked
                            ? 'bg-emerald-600 hover:bg-emerald-500'
                            : 'bg-red-600 hover:bg-red-500'
                        }`}
                      >
                        {busyId === c.id ? '...' : c.is_blocked ? 'Unblock' : 'Block'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!isLoading && totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-6">
            <button
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-sm disabled:opacity-40"
            >
              Prev
            </button>
            <span className="text-sm text-slate-400">
              Page {page} of {totalPages}
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-sm disabled:opacity-40"
            >
              Next
            </button>
          </div>
        )}
      </div>

      {(selected || isDetailLoading) && (
        <Modal title={selected ? (selected.name?.trim() || `Customer #${selected.id}`) : "Loading..."} onClose={() => setSelected(null)}>
          {isDetailLoading && <p className="text-slate-400 p-6">Loading...</p>}
          {selected && !isDetailLoading && (
            <div className="p-6 max-w-2xl max-h-[80vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold">
                  {selected.name?.trim() || `Customer #${selected.id}`}
                </h2>
                <span
                  className={`px-2 py-1 rounded-md text-xs font-medium ${
                    selected.is_blocked
                      ? 'bg-red-500/15 text-red-300'
                      : 'bg-emerald-500/15 text-emerald-300'
                  }`}
                >
                  {selected.is_blocked ? 'Blocked' : 'Active'}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm mb-6">
                <div>
                  <p className="text-slate-500">Phone</p>
                  <p className="text-slate-200">{selected.phone}</p>
                </div>
                <div>
                  <p className="text-slate-500">Joined</p>
                  <p className="text-slate-200">{new Date(selected.created_at).toLocaleDateString()}</p>
                </div>
                <div>
                  <p className="text-slate-500">Total Orders</p>
                  <p className="text-slate-200">{selected.total_orders}</p>
                </div>
                <div>
                  <p className="text-slate-500">Total Spent</p>
                  <p className="text-slate-200">{"\u20B9"}{selected.total_spent.toFixed(2)}</p>
                </div>
                {selected.wallet && (
                  <div>
                    <p className="text-slate-500">Wallet Balance</p>
                    <p className="text-slate-200">{"\u20B9"}{selected.wallet.balance.toFixed(2)}</p>
                  </div>
                )}
              </div>

              <h3 className="text-sm font-semibold text-slate-300 mb-2">Addresses</h3>
              {selected.addresses.length === 0 && (
                <p className="text-slate-500 text-sm mb-4">No addresses saved.</p>
              )}
              <div className="space-y-2 mb-6">
                {selected.addresses.map((a) => (
                  <div key={a.id} className="border border-slate-800 rounded-lg p-3 text-sm">
                    <p className="text-slate-200">{a.full_name} &middot; {a.phone}</p>
                    <p className="text-slate-400">
                      {a.line1}{a.line2 ? `, ${a.line2}` : ''}, {a.city}, {a.state} - {a.pincode}
                    </p>
                    {a.is_default && (
                      <span className="text-xs text-indigo-300">Default</span>
                    )}
                  </div>
                ))}
              </div>

              <h3 className="text-sm font-semibold text-slate-300 mb-2">Recent Orders</h3>
              {selected.orders.length === 0 && (
                <p className="text-slate-500 text-sm">No orders yet.</p>
              )}
              <div className="space-y-2">
                {selected.orders.slice(0, 10).map((o: any) => (
                  <div key={o.id} className="border border-slate-800 rounded-lg p-3 text-sm flex items-center justify-between">
                    <div>
                      <p className="text-slate-200">Order #{o.id}</p>
                      <p className="text-slate-500">{new Date(o.created_at).toLocaleDateString()}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-slate-200">{"\u20B9"}{o.total_amount}</p>
                      <p className="text-slate-500 text-xs">{o.status}</p>
                    </div>
                  </div>
                ))}
              </div>

              <div className="mt-6 flex justify-end">
                <button
                  disabled={busyId === selected.id}
                  onClick={() => handleToggleBlock(selected)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-50 ${
                    selected.is_blocked
                      ? 'bg-emerald-600 hover:bg-emerald-500'
                      : 'bg-red-600 hover:bg-red-500'
                  }`}
                >
                  {busyId === selected.id ? '...' : selected.is_blocked ? 'Unblock Customer' : 'Block Customer'}
                </button>
              </div>
            </div>
          )}
        </Modal>
      )}
    </Layout>
  )
}
