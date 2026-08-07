import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { getCart, addToCart as apiAddToCart, updateCartItem as apiUpdateCartItem, removeFromCart as apiRemoveFromCart } from '../api/cart'
import type { CartResponse } from '../types'
import { useAuth } from './AuthContext'

interface CartContextType {
  cart: CartResponse | null
  isLoading: boolean
  refreshCart: () => Promise<void>
  addItem: (productId: number, quantity: number) => Promise<void>
  updateItem: (itemId: number, quantity: number) => Promise<void>
  removeItem: (itemId: number) => Promise<void>
}

const CartContext = createContext<CartContextType | undefined>(undefined)

export function CartProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const [cart, setCart] = useState<CartResponse | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  async function refreshCart() {
    if (!user) {
      setCart(null)
      return
    }
    setIsLoading(true)
    try {
      const data = await getCart()
      setCart(data)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    refreshCart()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user])

  async function addItem(productId: number, quantity: number) {
    const data = await apiAddToCart(productId, quantity)
    setCart(data)
  }

  async function updateItem(itemId: number, quantity: number) {
    const data = await apiUpdateCartItem(itemId, quantity)
    setCart(data)
  }

  async function removeItem(itemId: number) {
    const data = await apiRemoveFromCart(itemId)
    setCart(data)
  }

  return (
    <CartContext.Provider value={{ cart, isLoading, refreshCart, addItem, updateItem, removeItem }}>
      {children}
    </CartContext.Provider>
  )
}

export function useCart() {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error('useCart must be used within CartProvider')
  return ctx
}
