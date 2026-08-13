import apiClient from './client'
import type {
  StockTransfer,
  OrdersResponse,
  PickingTask,
  PickingTaskItem,
  PackingTaskResponse,
  PackingTask,
  WarehouseDashboardStats,
  PickItemStatus,
} from '../types/warehouse'

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

// ---- Dashboard ----

export const getDashboard = () =>
  apiClient.get('/warehouse/dashboard').then((r) => r.data as WarehouseDashboardStats)

// ---- Orders ----

export const listWarehouseOrders = (params: { status?: string; page?: number; limit?: number }) =>
  apiClient.get('/warehouse/orders', { params }).then((r) => r.data as OrdersResponse)

export const acceptOrder = (orderId: number) =>
  apiClient
    .put(`/warehouse/orders/${orderId}/accept`)
    .then((r) => r.data as { success: boolean; order_id: number; status: string })

// ---- Picking ----

export const getPickingTask = (orderId: number) =>
  apiClient.get(`/warehouse/picking/${orderId}`).then((r) => r.data as PickingTask)

export const startPicking = (orderId: number) =>
  apiClient.put(`/warehouse/picking/${orderId}/start`).then((r) => r.data as PickingTask)

export const markPickItem = (
  itemId: number,
  data: { status: PickItemStatus; quantity_picked?: number; reason?: string }
) => apiClient.put(`/warehouse/picking/items/${itemId}`, data).then((r) => r.data as PickingTaskItem)

export const completePicking = (orderId: number) =>
  apiClient
    .put(`/warehouse/picking/${orderId}/complete`)
    .then((r) => r.data as { success: boolean; picking_task: PickingTask; packing_task: PackingTask })

// ---- Packing ----

export const getPackingTask = (orderId: number) =>
  apiClient.get(`/warehouse/packing/${orderId}`).then((r) => r.data as PackingTaskResponse)

export const startPacking = (orderId: number) =>
  apiClient.put(`/warehouse/packing/${orderId}/start`).then((r) => r.data as PackingTask)

export const completePacking = (orderId: number) =>
  apiClient
    .put(`/warehouse/packing/${orderId}/complete`)
    .then((r) => r.data as { success: boolean; packing_task: PackingTask; order_status: string })
