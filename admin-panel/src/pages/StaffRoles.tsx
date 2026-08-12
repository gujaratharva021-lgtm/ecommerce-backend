import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import { listAdminStaff, updateStaffRole } from '../api/admin'
import type { StaffMember } from '../types/admin'
import { ADMIN_ROLES } from '../types/admin'

export default function StaffRoles() {
  const [staff, setStaff] = useState<StaffMember[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<number | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listAdminStaff()
      setStaff(res.staff ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load staff.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function handleRoleChange(id: number, role: string) {
    setBusyId(id)
    try {
      await updateStaffRole(id, role)
      setStaff((prev) => prev.map((s) => (s.id === id ? { ...s, admin_role: role } : s)))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update role. You may not have permission for this action.')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Staff & Roles</h1>
          <p className="text-sm text-slate-400 mt-1">
            {staff.length} admin{staff.length !== 1 ? 's' : ''} &middot; role changes require Staff Management permission
          </p>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && staff.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No admin accounts found.
          </div>
        )}

        {!isLoading && staff.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Phone</th>
                  <th className="px-4 py-3 font-medium">Joined</th>
                  <th className="px-4 py-3 font-medium">Role</th>
                </tr>
              </thead>
              <tbody>
                {staff.map((s) => (
                  <tr key={s.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">{s.name?.trim() || `Admin #${s.id}`}</td>
                    <td className="px-4 py-3 text-slate-300">{s.phone}</td>
                    <td className="px-4 py-3 text-slate-400">
                      {new Date(s.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={s.admin_role ?? ''}
                        disabled={busyId === s.id}
                        onChange={(e) => handleRoleChange(s.id, e.target.value)}
                        className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
                      >
                        <option value="">Unassigned (full access)</option>
                        {ADMIN_ROLES.map((r) => (
                          <option key={r.value} value={r.value}>
                            {r.label}
                          </option>
                        ))}
                      </select>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </Layout>
  )
}
