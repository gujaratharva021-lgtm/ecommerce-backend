import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listProducts, listCategories } from '../api/products'
import { IMAGE_ORIGIN } from '../api/client'
import type { Product, Category } from '../types'

export default function Home() {
  const [products, setProducts] = useState<Product[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      listProducts({ sort: 'newest', limit: 8 }),
      listCategories(),
    ])
      .then(([productsRes, categoriesRes]) => {
        setProducts(productsRes.products ?? [])
        const cats = Array.isArray(categoriesRes) ? categoriesRes : (categoriesRes?.categories ?? [])
        setCategories(Array.isArray(cats) ? cats : [])
      })
      .finally(() => setIsLoading(false))
  }, [])

  function imageUrl(path: string) {
    if (!path) return ''
    return path.startsWith('http') ? path : `${IMAGE_ORIGIN}${path}`
  }

  return (
    <div>
      <section className="border-b border-line bg-gradient-to-b from-marigold/10 to-transparent">
        <div className="max-w-6xl mx-auto px-6 py-20 text-center">
          <p className="font-mono text-xs tracking-widest text-marigold uppercase mb-4">
            Fresh stock, daily
          </p>
          <h1 className="font-display text-5xl md:text-6xl font-600 mb-4 leading-tight">
            The market,
            <br />
            delivered to your door.
          </h1>
          <p className="text-ink/60 max-w-md mx-auto mb-8">
            Real prices, real stock, no markup games. Browse what's fresh today.
          </p>
          <Link
            to="/products"
            className="inline-block bg-ink text-paper font-medium px-8 py-3 rounded-lg hover:bg-marigold transition-colors"
          >
            Start shopping
          </Link>
        </div>
      </section>

      {categories.length > 0 && (
        <section className="max-w-6xl mx-auto px-6 py-14">
          <h2 className="font-display text-2xl font-600 mb-6">Shop by category</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4">
            {categories.map((c) => (
              <Link
                key={c.id}
                to={`/products?category_id=${c.id}`}
                className="group text-center"
              >
                <div className="aspect-square rounded-xl bg-line/40 overflow-hidden mb-2 border border-line group-hover:border-marigold transition-colors">
                  {c.image_url ? (
                    <img
                      src={imageUrl(c.image_url)}
                      alt={c.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-2xl">
                      🛒
                    </div>
                  )}
                </div>
                <p className="text-sm font-medium">{c.name}</p>
              </Link>
            ))}
          </div>
        </section>
      )}

      <section className="max-w-6xl mx-auto px-6 py-14">
        <div className="flex items-center justify-between mb-6">
          <h2 className="font-display text-2xl font-600">Just arrived</h2>
          <Link to="/products" className="text-sm font-medium text-marigold hover:underline">
            View all →
          </Link>
        </div>

        {isLoading ? (
          <p className="text-ink/50">Loading...</p>
        ) : products.length === 0 ? (
          <p className="text-ink/50">No products yet — check back soon.</p>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-5">
            {products.map((p) => (
              <Link
                key={p.id}
                to={`/products/${p.id}`}
                className="ticket-edge group border border-line rounded-xl overflow-hidden hover:shadow-lg transition-shadow bg-white/40"
              >
                <div className="aspect-square bg-line/30 overflow-hidden">
                  {p.image_url ? (
                    <img
                      src={imageUrl(p.image_url)}
                      alt={p.name}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-3xl">📦</div>
                  )}
                </div>
                <div className="p-3">
                  <p className="text-sm font-medium truncate">{p.name}</p>
                  <p className="font-mono text-marigold font-semibold mt-1">₹{p.price}</p>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

