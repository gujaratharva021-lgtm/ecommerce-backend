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

export const getOrderInvoice = (id: number) =>
  apiClient.get(`/orders/${id}/invoice`).then((r) => r.data)

// Downloads the invoice PDF and triggers a browser save - a plain <a href>
// can't be used because the request needs the Authorization header, so we
// fetch it as a blob and hand the browser a local object URL instead.
export const downloadOrderInvoicePDF = async (id: number) => {
  const res = await apiClient.get(`/orders/${id}/invoice/pdf`, { responseType: 'blob' })
  const url = window.URL.createObjectURL(new Blob([res.data]))
  const link = document.createElement('a')
  link.href = url
  const disposition = res.headers['content-disposition'] as string | undefined
  const match = disposition?.match(/filename="(.+)"/)
  link.download = match?.[1] ?? `invoice-order-${id}.pdf`
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}

export const createPaymentOrder = (id: number) =>
  apiClient.post<CreatePaymentOrderResponse>(`/orders/${id}/payment`, {}).then((r) => r.data)

export const verifyPayment = (id: number, data: VerifyPaymentRequest) =>
  apiClient.post(`/orders/${id}/payment/verify`, data).then((r) => r.data)
