import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const navItems = [
  { to: '/dashboard', label: 'Overview', icon: '\u25C6' },
  { to: '/customers', label: 'Customers', icon: '\u25C7' },
  { to: '/inventory', label: 'Inventory Overview', icon: '\u25A6' },
  { to: '/staff-roles', label: 'Staff & Roles', icon: '\u25C8' },
  { to: '/settings', label: 'Settings', icon: '\u2699' },
  { to: '/products', label: 'Products', icon: '\u25A3' },
  { to: '/categories', label: 'Categories', icon: '\u25A4' },
  { to: '/orders', label: 'Orders', icon: '\u25A5' },
  { to: '/coupons', label: 'Coupons', icon: '\u25A7' },
  { to: '/delivery-partners', label: 'Delivery Partners', icon: '\u25B2' },
  { to: '/warehouses', label: 'Warehouses', icon: '\u25A0' },
  { to: '/warehouse-staff', label: 'Warehouse Staff', icon: '\u25AB' },
  { to: '/stock-transfers', label: 'Stock Transfers', icon: '\u21C4' },
  { to: '/returns', label: 'Returns', icon: '\u21BA' },
  { to: '/analytics', label: 'Analytics', icon: '\u25C9' },
  { to: '/wallet-credit', label: 'Wallet Credit', icon: '\u25CF' },
]

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth()

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex">
      <aside className="w-60 shrink-0 border-r border-slate-800 bg-slate-900 flex flex-col">
        <div className="px-6 py-5 border-b border-slate-800">
          <p className="text-sm font-semibold tracking-wide text-slate-100">
            Ecommerce Admin
          </p>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-indigo-500/15 text-indigo-300'
                    : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800'
                }`
              }
            >
              <span className="text-xs opacity-70">{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="px-4 py-4 border-t border-slate-800">
          <p className="text-xs text-slate-500 mb-2 truncate">
            {user?.phone} &middot; {user?.role}
          </p>
          <button
            onClick={logout}
            className="w-full text-left text-sm px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
          >
            Log out
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-y-auto">{children}</main>
    </div>
  )
}

