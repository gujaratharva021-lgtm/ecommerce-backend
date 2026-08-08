import { apiClient } from './client'
import type {
  Order,
  OrderListResponse,
  OrderTracking,
  CheckoutRequest,
  CreatePaymentOrderResponse,
  VerifyPaymentRequest,
  ReturnRequestBody,
} from '../types'

export const checkout = (data: CheckoutRequest) =>
  apiClient.post<Order>('/orders/checkout', data).then((r) => r.data)

export const listOrders = () =>
  apiClient.get<OrderListResponse>('/orders').then((r) => r.data)

export const getOrder = (id: number) =>
  apiClient.get<Order>(`/orders/${id}`).then((r) => r.data)

export const getOrderTracking = (id: number) =>
  apiClient
    .get<{ message?: string; tracking: OrderTracking | null }>(`/orders/${id}/tracking`)
    .then((r) => r.data.tracking)

export const cancelOrder = (id: number) =>
  apiClient.put<Order>(`/orders/${id}/cancel`, {}).then((r) => r.data)

export const requestReturn = (id: number, data: ReturnRequestBody) =>
  apiClient.post(`/orders/${id}/return`, data).then((r) => r.data)

export const createPaymentOrder = (id: number) =>
  apiClient.post<CreatePaymentOrderResponse>(`/orders/${id}/payment`, {}).then((r) => r.data)

export const verifyPayment = (id: number, data: VerifyPaymentRequest) =>
  apiClient.post(`/orders/${id}/payment/verify`, data).then((r) => r.data)
