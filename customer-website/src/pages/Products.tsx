import { useEffect, useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { listProducts, listCategories } from '../api/products'
import { IMAGE_ORIGIN } from '../api/client'
import type { Product, Category } from '../types'

const SORT_OPTIONS = [
  { value: '', label: 'Relevance' },
  { value: 'newest', label: 'Newest' },
  { value: 'price_asc', label: 'Price: Low to High' },
  { value: 'price_desc', label: 'Price: High to Low' },
  { value: 'name_asc', label: 'Name: A-Z' },
]

export default function Products() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [products, setProducts] = useState<Product[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [totalPages, setTotalPages] = useState(1)

  const search = searchParams.get('search') ?? ''
  const categoryId = searchParams.get('category_id') ?? ''
  const sort = searchParams.get('sort') ?? ''
  const page = parseInt(searchParams.get('page') ?? '1', 10)

  useEffect(() => {
    listCategories().then((res: any) => {
      const cats = Array.isArray(res) ? res : (res?.categories ?? [])
      setCategories(Array.isArray(cats) ? cats : [])
    })
  }, [])

  useEffect(() => {
    setIsLoading(true)
    listProducts({
      search: search || undefined,
      category_id: categoryId ? parseInt(categoryId, 10) : undefined,
      sort: (sort as any) || undefined,
      page,
      limit: 20,
    })
      .then((res) => {
        setProducts(res.products ?? [])
        setTotalPages(res.total_pages ?? 1)
      })
      .finally(() => setIsLoading(false))
  }, [search, categoryId, sort, page])

  function updateParam(key: string, value: string) {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== 'page') next.delete('page')
    setSearchParams(next)
  }

  function imageUrl(path: string) {
    if (!path) return ''
    return path.startsWith('http') ? path : `${IMAGE_ORIGIN}${path}`
  }

  return (
    <div className="max-w-6xl mx-auto px-6 py-10">
      <h1 className="font-display text-3xl font-600 mb-6">Shop</h1>

      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <input
          type="text"
          defaultValue={search}
          onKeyDown={(e) => {
            if (e.key === 'Enter') updateParam('search', (e.target as HTMLInputElement).value)
          }}
          placeholder="Search products..."
          className="flex-1 border border-line rounded-lg px-4 py-2.5 outline-none focus:border-ink"
        />

        <select
          value={categoryId}
          onChange={(e) => updateParam('category_id', e.target.value)}
          className="border border-line rounded-lg px-3 py-2.5 outline-none focus:border-ink bg-paper"
        >
          <option value="">All categories</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>

        <select
          value={sort}
          onChange={(e) => updateParam('sort', e.target.value)}
          className="border border-line rounded-lg px-3 py-2.5 outline-none focus:border-ink bg-paper"
        >
          {SORT_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <p className="text-ink/50">Loading...</p>
      ) : products.length === 0 ? (
        <div className="border border-dashed border-line rounded-xl p-16 text-center text-ink/50">
          No products found. Try a different search or category.
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-5">
            {products.map((p) => {
              const inStock = p.inventories?.some((i) => i.in_stock && i.stock > 0) ?? true
              return (
                <Link
                  key={p.id}
                  to={`/products/${p.id}`}
                  className="ticket-edge group border border-line rounded-xl overflow-hidden hover:shadow-lg transition-shadow bg-white/40"
                >
                  <div className="aspect-square bg-line/30 overflow-hidden relative">
                    {p.image_url ? (
                      <img
                        src={imageUrl(p.image_url)}
                        alt={p.name}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-3xl">📦</div>
                    )}
                    {!inStock && (
                      <div className="absolute inset-0 bg-paper/70 flex items-center justify-center">
                        <span className="text-xs font-medium bg-ink text-paper px-2 py-1 rounded">
                          Out of stock
                        </span>
                      </div>
                    )}
                  </div>
                  <div className="p-3">
                    <p className="text-sm font-medium truncate">{p.name}</p>
                    <p className="font-mono text-marigold font-semibold mt-1">₹{p.price}</p>
                  </div>
                </Link>
              )
            })}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-10">
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <button
                  key={p}
                  onClick={() => updateParam('page', String(p))}
                  className={`w-9 h-9 rounded-lg text-sm font-medium transition-colors ${
                    p === page
                      ? 'bg-ink text-paper'
                      : 'border border-line hover:border-ink'
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}

