import { apiClient } from './client'
import type { Product, ProductListResponse, ProductListQuery, Review } from '../types'

export const listProducts = (params?: ProductListQuery) =>
  apiClient.get<ProductListResponse>('/products', { params }).then((r) => r.data)

export const getProduct = (id: number) =>
  apiClient.get<Product>(`/products/${id}`).then((r) => r.data)

export const listCategories = () =>
  apiClient.get('/categories').then((r) => r.data)

export const getProductReviews = (productId: number) =>
  apiClient.get<Review[]>(`/products/${productId}/reviews`).then((r) => r.data)

export const upsertReview = (productId: number, rating: number, comment: string) =>
  apiClient.post(`/products/${productId}/reviews`, { rating, comment }).then((r) => r.data)

export const deleteReview = (productId: number) =>
  apiClient.delete(`/products/${productId}/reviews`).then((r) => r.data)
