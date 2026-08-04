import type { ReactNode } from 'react'
import { useAuth } from '../context/AuthContext'

export default function Layout({ children }: { children: ReactNode }) {
  const { staff, logout } = useAuth()

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-900 px-6 py-4 flex items-center justify-between">
        <div>
          <p className="text-sm font-semibold">Warehouse Staff</p>
          <p className="text-xs text-slate-400">
            {staff?.name} &middot; {staff?.phone}
            {staff?.warehouse?.name ? ` \u00b7 ${staff.warehouse.name}` : ''}
          </p>
        </div>
        <button
          onClick={logout}
          className="text-xs px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          Log out
        </button>
      </header>
      <main>{children}</main>
    </div>
  )
}
