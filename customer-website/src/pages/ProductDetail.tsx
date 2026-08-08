import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getProduct, getProductReviews, upsertReview } from '../api/products'
import { getWishlist, addToWishlist, removeFromWishlist } from '../api/misc'
import { IMAGE_ORIGIN } from '../api/client'
import { useCart } from '../context/CartContext'
import { useAuth } from '../context/AuthContext'
import type { Product, Review } from '../types'

export default function ProductDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { addItem } = useCart()
  const { user } = useAuth()

  const [product, setProduct] = useState<Product | null>(null)
  const [reviews, setReviews] = useState<Review[]>([])
  const [quantity, setQuantity] = useState(1)
  const [isAdding, setIsAdding] = useState(false)
  const [addedMsg, setAddedMsg] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [myRating, setMyRating] = useState(5)
  const [myComment, setMyComment] = useState('')
  const [isSubmittingReview, setIsSubmittingReview] = useState(false)

  const [isWishlisted, setIsWishlisted] = useState(false)
  const [isTogglingWishlist, setIsTogglingWishlist] = useState(false)

  useEffect(() => {
    if (!id) return
    const productId = parseInt(id, 10)
    getProduct(productId).then(setProduct)
    getProductReviews(productId).then((r) => setReviews(Array.isArray(r) ? r : []))
  }, [id])

  useEffect(() => {
    if (!user || !id) {
      setIsWishlisted(false)
      return
    }
    const productId = parseInt(id, 10)
    getWishlist()
      .then((items) => setIsWishlisted(items.some((w) => w.product_id === productId)))
      .catch(() => setIsWishlisted(false))
  }, [user, id])

  async function handleToggleWishlist() {
    if (!user) {
      navigate('/login')
      return
    }
    if (!product) return
    setIsTogglingWishlist(true)
    try {
      if (isWishlisted) {
        await removeFromWishlist(product.id)
        setIsWishlisted(false)
      } else {
        await addToWishlist(product.id)
        setIsWishlisted(true)
      }
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to update wishlist.')
    } finally {
      setIsTogglingWishlist(false)
    }
  }

  function imageUrl(path: string) {
    if (!path) return ''
    return path.startsWith('http') ? path : `${IMAGE_ORIGIN}${path}`
  }

  async function handleAddToCart() {
    if (!user) {
      navigate('/login')
      return
    }
    if (!product) return
    setIsAdding(true)
    setError(null)
    try {
      await addItem(product.id, quantity)
      setAddedMsg(true)
      setTimeout(() => setAddedMsg(false), 2000)
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to add to cart.')
    } finally {
      setIsAdding(false)
    }
  }

  async function handleSubmitReview(e: React.FormEvent) {
    e.preventDefault()
    if (!product) return
    setIsSubmittingReview(true)
    try {
      await upsertReview(product.id, myRating, myComment)
      const updated = await getProductReviews(product.id)
      setReviews(Array.isArray(updated) ? updated : [])
      setMyComment('')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to submit review.')
    } finally {
      setIsSubmittingReview(false)
    }
  }

  if (!product) {
    return <div className="max-w-6xl mx-auto px-6 py-16 text-ink/50">Loading...</div>
  }

  const totalStock = product.inventories?.reduce((sum, i) => sum + (i.in_stock ? i.stock : 0), 0) ?? 0
  const inStock = totalStock > 0

  const avgRating =
    reviews.length > 0 ? reviews.reduce((s, r) => s + r.rating, 0) / reviews.length : null

  return (
    <div className="max-w-6xl mx-auto px-6 py-10">
      <div className="grid md:grid-cols-2 gap-10">
        <div className="aspect-square rounded-xl overflow-hidden bg-line/30 border border-line">
          {product.image_url ? (
            <img
              src={imageUrl(product.image_url)}
              alt={product.name}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-5xl">📦</div>
          )}
        </div>

        <div>
          {product.category && (
            <p className="text-xs font-mono uppercase tracking-widest text-marigold mb-2">
              {product.category.name}
            </p>
          )}
          <div className="flex items-start justify-between gap-3 mb-3">
            <h1 className="font-display text-3xl font-600">{product.name}</h1>
            <button
              onClick={handleToggleWishlist}
              disabled={isTogglingWishlist}
              aria-label={isWishlisted ? 'Remove from wishlist' : 'Add to wishlist'}
              className="shrink-0 text-2xl leading-none mt-1 disabled:opacity-50 transition-transform hover:scale-110"
            >
              {isWishlisted ? '❤️' : '🤍'}
            </button>
          </div>

          {avgRating !== null && (
            <p className="text-sm text-ink/60 mb-3">
              ⭐ {avgRating.toFixed(1)} ({reviews.length} review{reviews.length !== 1 ? 's' : ''})
            </p>
          )}

          <p className="font-mono text-3xl font-semibold text-marigold mb-4">₹{product.price}</p>

          <p className="text-ink/70 mb-6 leading-relaxed">{product.description}</p>

          <p className={`text-sm font-medium mb-6 ${inStock ? 'text-leaf' : 'text-clay'}`}>
            {inStock ? `In stock (${totalStock} available)` : 'Out of stock'}
          </p>

          {inStock && (
            <div className="flex items-center gap-4 mb-4">
              <div className="flex items-center border border-line rounded-lg">
                <button
                  onClick={() => setQuantity((q) => Math.max(1, q - 1))}
                  className="w-10 h-10 flex items-center justify-center hover:bg-line/40"
                >
                  −
                </button>
                <span className="w-10 text-center font-mono">{quantity}</span>
                <button
                  onClick={() => setQuantity((q) => Math.min(totalStock, q + 1))}
                  className="w-10 h-10 flex items-center justify-center hover:bg-line/40"
                >
                  +
                </button>
              </div>

              <button
                onClick={handleAddToCart}
                disabled={isAdding}
                className="flex-1 bg-ink text-paper font-medium py-2.5 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
              >
                {isAdding ? 'Adding...' : addedMsg ? 'Added ✓' : 'Add to cart'}
              </button>
            </div>
          )}

          {error && <p className="text-clay text-sm">{error}</p>}
        </div>
      </div>

      <section className="mt-16 max-w-2xl">
        <h2 className="font-display text-2xl font-600 mb-6">Reviews</h2>

        {user && (
          <form onSubmit={handleSubmitReview} className="border border-line rounded-xl p-5 mb-8">
            <p className="text-sm font-medium mb-2">Leave a review</p>
            <div className="flex gap-1 mb-3">
              {[1, 2, 3, 4, 5].map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => setMyRating(n)}
                  className={`text-2xl ${n <= myRating ? 'opacity-100' : 'opacity-25'}`}
                >
                  ⭐
                </button>
              ))}
            </div>
            <textarea
              value={myComment}
              onChange={(e) => setMyComment(e.target.value)}
              placeholder="What did you think?"
              className="w-full border border-line rounded-lg px-3 py-2 outline-none focus:border-ink mb-3 resize-none"
              rows={3}
            />
            <button
              type="submit"
              disabled={isSubmittingReview}
              className="bg-ink text-paper text-sm font-medium px-4 py-2 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
            >
              {isSubmittingReview ? 'Submitting...' : 'Submit review'}
            </button>
          </form>
        )}

        {reviews.length === 0 ? (
          <p className="text-ink/50 text-sm">No reviews yet. Be the first to share your thoughts.</p>
        ) : (
          <div className="space-y-4">
            {reviews.map((r) => (
              <div key={r.id} className="border-b border-line pb-4">
                <p className="text-sm mb-1">{'⭐'.repeat(r.rating)}</p>
                {r.comment && <p className="text-ink/70 text-sm">{r.comment}</p>}
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
