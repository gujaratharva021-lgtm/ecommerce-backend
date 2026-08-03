import { type ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth()
  const { cart } = useCart()
  const navigate = useNavigate()

  const itemCount = cart?.total_items ?? 0

  function handleLogout() {
    logout()
    navigate('/')
  }

  return (
    <div className="min-h-screen flex flex-col bg-paper text-ink">
      <header className="border-b border-line sticky top-0 bg-paper/95 backdrop-blur z-10">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between gap-6">
          <Link to="/" className="font-display text-2xl font-600 tracking-tight shrink-0">
            Bazaar<span className="text-marigold">.</span>
          </Link>

          <nav className="hidden md:flex items-center gap-6 text-sm font-medium">
            <Link to="/products" className="hover:text-marigold transition-colors">
              Shop
            </Link>
            {user && (
              <Link to="/orders" className="hover:text-marigold transition-colors">
                Orders
              </Link>
            )}
            {user && (
              <Link to="/wallet" className="hover:text-marigold transition-colors">
                Wallet
              </Link>
            )}
            {user && (
              <Link to="/wishlist" className="hover:text-marigold transition-colors">
                Wishlist
              </Link>
            )}
          </nav>

          <div className="flex items-center gap-4">
            <Link
              to="/cart"
              className="relative flex items-center gap-1.5 text-sm font-medium hover:text-marigold transition-colors"
            >
              <span>Cart</span>
              {itemCount > 0 && (
                <span className="inline-flex items-center justify-center w-5 h-5 rounded-full bg-marigold text-white text-xs font-mono font-semibold">
                  {itemCount}
                </span>
              )}
            </Link>

            {user ? (
              <div className="flex items-center gap-3 text-sm">
                <Link
                  to="/profile"
                  className="hidden sm:inline text-ink/60 hover:text-marigold transition-colors"
                >
                  Hi, {user.name || user.phone}
                </Link>
                <button
                  onClick={handleLogout}
                  className="px-3 py-1.5 rounded-full border border-line hover:border-ink transition-colors"
                >
                  Log out
                </button>
              </div>
            ) : (
              <Link
                to="/login"
                className="px-4 py-1.5 rounded-full bg-ink text-paper text-sm font-medium hover:bg-marigold transition-colors"
              >
                Log in
              </Link>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1">{children}</main>

      <footer className="border-t border-line mt-16">
        <div className="max-w-6xl mx-auto px-6 py-8 text-sm text-ink/50 flex items-center justify-between">
          <span>© {new Date().getFullYear()} Bazaar. All rights reserved.</span>
          <span className="font-mono text-xs">Fresh picks, fast delivery.</span>
        </div>
      </footer>
    </div>
  )
}
