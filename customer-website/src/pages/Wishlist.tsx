import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getWishlist, removeFromWishlist } from '../api/misc'
import { IMAGE_ORIGIN } from '../api/client'
import type { Wishlist as WishlistItem } from '../types'

export default function Wishlist() {
  const [items, setItems] = useState<WishlistItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [removingId, setRemovingId] = useState<number | null>(null)

  function load() {
    setIsLoading(true)
    getWishlist()
      .then(setItems)
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  function imageUrl(path?: string) {
    if (!path) return ''
    return path.startsWith('http') ? path : `${IMAGE_ORIGIN}${path}`
  }

  async function handleRemove(productId: number) {
    setRemovingId(productId)
    try {
      await removeFromWishlist(productId)
      setItems((prev) => prev.filter((i) => i.product_id !== productId))
    } finally {
      setRemovingId(null)
    }
  }

  return (
    <div className="max-w-4xl mx-auto px-6 py-10">
      <h1 className="font-display text-3xl font-600 mb-6">Your wishlist</h1>

      {isLoading ? (
        <p className="text-ink/50">Loading...</p>
      ) : items.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-4xl mb-3">🤍</p>
          <p className="text-ink/50 text-sm mb-4">Nothing saved yet.</p>
          <Link
            to="/products"
            className="text-sm font-medium text-marigold hover:underline"
          >
            Browse products
          </Link>
        </div>
      ) : (
        <div className="grid sm:grid-cols-2 md:grid-cols-3 gap-4">
          {items.map((item) => (
            <div key={item.id} className="border border-line rounded-xl overflow-hidden group">
              <Link to={`/products/${item.product_id}`}>
                <div className="aspect-square bg-line/30">
                  {item.product?.image_url ? (
                    <img
                      src={imageUrl(item.product.image_url)}
                      alt={item.product.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-3xl">📦</div>
                  )}
                </div>
              </Link>
              <div className="p-3">
                <Link to={`/products/${item.product_id}`}>
                  <p className="text-sm font-medium truncate hover:text-marigold transition-colors">
                    {item.product?.name ?? 'Product'}
                  </p>
                </Link>
                {item.product?.price != null && (
                  <p className="font-mono text-sm text-marigold mt-1">₹{item.product.price}</p>
                )}
                <button
                  onClick={() => handleRemove(item.product_id)}
                  disabled={removingId === item.product_id}
                  className="text-xs text-clay hover:underline mt-2 disabled:opacity-50"
                >
                  {removingId === item.product_id ? 'Removing...' : 'Remove'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
