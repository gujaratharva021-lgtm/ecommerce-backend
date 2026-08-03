import { apiClient } from './client'
import type { Coupon, WalletResponse, Wishlist } from '../types'

export const validateCoupon = (code: string, orderAmount: number) =>
  apiClient.post<{ coupon: Coupon; discount_amount: number }>('/coupons/validate', {
    code,
    order_amount: orderAmount,
  }).then((r) => r.data)

export const getWallet = () => apiClient.get<WalletResponse>('/wallet').then((r) => r.data)

export const getWishlist = () => apiClient.get<Wishlist[]>('/wishlist').then((r) => r.data)

export const addToWishlist = (productId: number) =>
  apiClient.post<Wishlist>('/wishlist', { product_id: productId }).then((r) => r.data)

export const removeFromWishlist = (productId: number) =>
  apiClient.delete(`/wishlist/${productId}`).then((r) => r.data)
