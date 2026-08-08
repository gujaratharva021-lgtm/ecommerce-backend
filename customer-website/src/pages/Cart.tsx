import { Link, useNavigate } from 'react-router-dom'
import { useCart } from '../context/CartContext'
import { IMAGE_ORIGIN } from '../api/client'
import { useState } from 'react'

export default function Cart() {
  const { cart, updateItem, removeItem } = useCart()
  const navigate = useNavigate()
  const [busyId, setBusyId] = useState<number | null>(null)

  function imageUrl(path: string) {
    if (!path) return ''
    return path.startsWith('http') ? path : `${IMAGE_ORIGIN}${path}`
  }

  async function handleQuantityChange(itemId: number, quantity: number) {
    if (quantity < 1) return
    setBusyId(itemId)
    try {
      await updateItem(itemId, quantity)
    } finally {
      setBusyId(null)
    }
  }

  async function handleRemove(itemId: number) {
    setBusyId(itemId)
    try {
      await removeItem(itemId)
    } finally {
      setBusyId(null)
    }
  }

  if (!cart || cart.items.length === 0) {
    return (
      <div className="max-w-2xl mx-auto px-6 py-20 text-center">
        <h1 className="font-display text-3xl font-600 mb-3">Your cart is empty</h1>
        <p className="text-ink/60 mb-8">Add something fresh to get started.</p>
        <Link
          to="/products"
          className="inline-block bg-ink text-paper font-medium px-6 py-3 rounded-lg hover:bg-marigold transition-colors"
        >
          Browse products
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-6 py-10">
      <h1 className="font-display text-3xl font-600 mb-8">Your cart</h1>

      <div className="space-y-4 mb-8">
        {cart.items.map((item) => (
          <div
            key={item.id}
            className="flex items-center gap-4 border border-line rounded-xl p-4"
          >
            <div className="w-20 h-20 rounded-lg bg-line/30 overflow-hidden shrink-0">
              {item.product?.image_url ? (
                <img
                  src={imageUrl(item.product.image_url)}
                  alt={item.product.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-xl">📦</div>
              )}
            </div>

            <div className="flex-1 min-w-0">
              <p className="font-medium truncate">{item.product?.name}</p>
              <p className="font-mono text-marigold text-sm mt-1">₹{item.product?.price}</p>
            </div>

            <div className="flex items-center border border-line rounded-lg">
              <button
                onClick={() => handleQuantityChange(item.id, item.quantity - 1)}
                disabled={busyId === item.id}
                className="w-8 h-8 flex items-center justify-center hover:bg-line/40 disabled:opacity-50"
              >
                −
              </button>
              <span className="w-8 text-center font-mono text-sm">{item.quantity}</span>
              <button
                onClick={() => handleQuantityChange(item.id, item.quantity + 1)}
                disabled={busyId === item.id}
                className="w-8 h-8 flex items-center justify-center hover:bg-line/40 disabled:opacity-50"
              >
                +
              </button>
            </div>

            <p className="font-mono font-semibold w-20 text-right">
              ₹{((item.product?.price ?? 0) * item.quantity).toFixed(2)}
            </p>

            <button
              onClick={() => handleRemove(item.id)}
              disabled={busyId === item.id}
              className="text-ink/40 hover:text-clay transition-colors disabled:opacity-50"
              aria-label="Remove item"
            >
              ✕
            </button>
          </div>
        ))}
      </div>

      <div className="border-t border-line pt-6 flex items-center justify-between">
        <div>
          <p className="text-sm text-ink/60">
            {cart.total_items} item{cart.total_items !== 1 ? 's' : ''}
          </p>
          <p className="font-mono text-2xl font-semibold">₹{cart.total_amount.toFixed(2)}</p>
        </div>
        <button
          onClick={() => navigate('/checkout')}
          className="bg-ink text-paper font-medium px-8 py-3 rounded-lg hover:bg-marigold transition-colors"
        >
          Checkout →
        </button>
      </div>
    </div>
  )
}
