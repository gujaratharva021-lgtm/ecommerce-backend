import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import Modal from '../components/Modal'
import { listCoupons, createCoupon, updateCouponStatus } from '../api/admin'
import type { Coupon } from '../types/admin'

const emptyForm = {
  code: '',
  discount_type: 'percentage' as 'percentage' | 'flat',
  discount_value: '',
  min_order_amount: '',
  usage_limit: '',
  expiry_date: '',
}

export default function Coupons() {
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const [form, setForm] = useState(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listCoupons()
      setCoupons(res.coupons ?? res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load coupons.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)
    if (!form.code.trim()) {
      setFormError('Coupon code is required.')
      return
    }
    if (!form.discount_value || parseFloat(form.discount_value) <= 0) {
      setFormError('Discount value must be greater than 0.')
      return
    }
    if (!form.expiry_date) {
      setFormError('Expiry date is required.')
      return
    }
    setIsSaving(true)
    try {
      await createCoupon({
        code: form.code.trim().toUpperCase(),
        discount_type: form.discount_type,
        discount_value: parseFloat(form.discount_value),
        min_order_amount: form.min_order_amount ? parseFloat(form.min_order_amount) : 0,
        usage_limit: form.usage_limit ? parseInt(form.usage_limit, 10) : 1,
        expiry_date: form.expiry_date,
      })
      setShowCreate(false)
      setForm(emptyForm)
      load()
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to create coupon.')
    } finally {
      setIsSaving(false)
    }
  }

  async function handleToggleStatus(c: Coupon) {
    try {
      await updateCouponStatus(c.id, !c.is_active)
      setCoupons((prev) =>
        prev.map((x) => (x.id === c.id ? { ...x, is_active: !c.is_active } : x))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update coupon status.')
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Coupons</h1>
            <p className="text-sm text-slate-400 mt-1">
              {coupons.length} coupon{coupons.length !== 1 ? 's' : ''}
            </p>
          </div>
          <button
            onClick={() => setShowCreate(true)}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            + New coupon
          </button>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && coupons.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No coupons yet.
          </div>
        )}

        {!isLoading && coupons.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Code</th>
                  <th className="px-4 py-3 font-medium">Discount</th>
                  <th className="px-4 py-3 font-medium">Min order</th>
                  <th className="px-4 py-3 font-medium">Usage</th>
                  <th className="px-4 py-3 font-medium">Expires</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {coupons.map((c) => (
                  <tr key={c.id} className="border-t border-slate-800">
                    <td className="px-4 py-3 font-mono">{c.code}</td>
                    <td className="px-4 py-3">
                      {c.discount_type === 'percentage'
                        ? `${c.discount_value}%`
                        : `₹${c.discount_value}`}
                    </td>
                    <td className="px-4 py-3">₹{c.min_order_amount}</td>
                    <td className="px-4 py-3">
                      {c.used_count}/{c.usage_limit}
                    </td>
                    <td className="px-4 py-3">
                      {new Date(c.expiry_date).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-1 rounded-md text-xs font-medium ${
                          c.is_active
                            ? 'bg-emerald-500/15 text-emerald-300'
                            : 'bg-slate-700 text-slate-300'
                        }`}
                      >
                        {c.is_active ? 'active' : 'inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => handleToggleStatus(c)}
                        className="text-indigo-400 hover:text-indigo-300 text-xs"
                      >
                        {c.is_active ? 'Deactivate' : 'Activate'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showCreate && (
        <Modal title="New coupon" onClose={() => setShowCreate(false)}>
          <form onSubmit={handleCreate} className="space-y-3">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Coupon code</label>
              <input
                value={form.code}
                onChange={(e) => setForm({ ...form, code: e.target.value })}
                placeholder="SAVE20"
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm font-mono"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Discount type</label>
                <select
                  value={form.discount_type}
                  onChange={(e) =>
                    setForm({ ...form, discount_type: e.target.value as 'percentage' | 'flat' })
                  }
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                >
                  <option value="percentage">Percentage</option>
                  <option value="flat">Flat amount</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-slate-400 block mb-1">
                  {form.discount_type === 'percentage' ? 'Discount %' : 'Discount ₹'}
                </label>
                <input
                  type="number"
                  value={form.discount_value}
                  onChange={(e) => setForm({ ...form, discount_value: e.target.value })}
                  placeholder="20"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Min order (₹)</label>
                <input
                  type="number"
                  value={form.min_order_amount}
                  onChange={(e) => setForm({ ...form, min_order_amount: e.target.value })}
                  placeholder="0"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="text-xs text-slate-400 block mb-1">Usage limit</label>
                <input
                  type="number"
                  value={form.usage_limit}
                  onChange={(e) => setForm({ ...form, usage_limit: e.target.value })}
                  placeholder="1"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Expiry date</label>
              <input
                type="date"
                value={form.expiry_date}
                onChange={(e) => setForm({ ...form, expiry_date: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>

            {formError && <p className="text-red-400 text-xs">{formError}</p>}

            <button
              type="submit"
              disabled={isSaving}
              className="w-full py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors mt-2"
            >
              {isSaving ? 'Saving...' : 'Create coupon'}
            </button>
          </form>
        </Modal>
      )}
    </Layout>
  )
}
