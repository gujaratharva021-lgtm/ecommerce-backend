import apiClient from './client'
import type { StockTransfer } from '../types/warehouse'
export const listMyStockTransfers = () =>
  apiClient.get('/warehouse/stock-transfers').then((r) => r.data as { stock_transfers: StockTransfer[] })
export const requestStockTransfer = (data: {
  product_id: number
  to_warehouse_id: number
  quantity: number
}) => apiClient.post('/warehouse/stock-transfers', data).then((r) => r.data)
export const receiveStockTransfer = (id: number) =>
  apiClient.put(`/warehouse/stock-transfers/${id}/receive`).then((r) => r.data)
export const approveStockTransfer = (id: number) =>
  apiClient.put(`/warehouse/stock-transfers/${id}/approve`).then((r) => r.data)
export const rejectStockTransfer = (id: number) =>
  apiClient.put(`/warehouse/stock-transfers/${id}/reject`).then((r) => r.data)
export const listProducts = (params?: Record<string, any>) =>
  apiClient.get('/products', { params }).then((r) => r.data)
