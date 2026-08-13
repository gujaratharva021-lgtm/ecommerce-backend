import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const navItems = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/orders', label: 'Orders' },
  { to: '/stock-transfers', label: 'Stock Transfers' },
  { to: '/locations', label: 'Locations' },
  { to: '/stock-operations', label: 'Stock Operations' },
  { to: '/receiving', label: 'Receiving' },
  { to: '/batches', label: 'Batch & Expiry' },
  { to: '/exceptions', label: 'Exceptions' },
  { to: '/performance', label: 'Performance' },
  { to: '/staff', label: 'Staff' },
]

export default function Layout({ children }: { children: ReactNode }) {
  const { staff, logout } = useAuth()

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex">
      <aside className="w-56 shrink-0 border-r border-slate-800 bg-slate-900 flex flex-col">
        <div className="px-5 py-5 border-b border-slate-800">
          <p className="text-sm font-semibold">Warehouse Panel</p>
          <p className="text-xs text-slate-400 mt-1">
            {staff?.warehouse?.name ?? `Warehouse #${staff?.warehouse_id ?? ''}`}
          </p>
        </div>
        <nav className="flex-1 px-2 py-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `block px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-indigo-500/15 text-indigo-300 font-medium'
                    : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="px-4 py-4 border-t border-slate-800">
          <p className="text-xs font-medium text-slate-200">{staff?.name}</p>
          <p className="text-xs text-slate-500 mb-3">{staff?.phone}</p>
          <button
            onClick={logout}
            className="w-full text-xs px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
          >
            Log out
          </button>
        </div>
      </aside>
      <main className="flex-1 min-w-0">{children}</main>
    </div>
  )
}
