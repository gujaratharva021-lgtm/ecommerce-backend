import { apiClient } from './client'
import type { CartResponse } from '../types'

export const getCart = () => apiClient.get<CartResponse>('/cart').then((r) => r.data)

export const addToCart = (productId: number, quantity: number) =>
  apiClient.post<CartResponse>('/cart', { product_id: productId, quantity }).then((r) => r.data)

export const updateCartItem = (itemId: number, quantity: number) =>
  apiClient.put<CartResponse>(`/cart/${itemId}`, { quantity }).then((r) => r.data)

export const removeFromCart = (itemId: number) =>
  apiClient.delete<CartResponse>(`/cart/${itemId}`).then((r) => r.data)
