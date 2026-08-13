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
  ScanResult,
  ExceptionsResponse,
  WarehouseException,
  StaffPerformanceRow,
  WarehouseZone,
  WarehouseRack,
  WarehouseBin,
  Inventory,
  StockMovementsResponse,
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


// ---- Exceptions ----

export const listExceptions = (params: {
  status?: string
  priority?: string
  type?: string
  page?: number
  limit?: number
}) => apiClient.get('/warehouse/exceptions', { params }).then((r) => r.data as ExceptionsResponse)

export const updateException = (id: number, data: { status: string; resolution?: string }) =>
  apiClient.put(`/warehouse/exceptions/${id}`, data).then((r) => r.data as WarehouseException)

// ---- Staff performance ----

export const getStaffPerformance = () =>
  apiClient
    .get('/warehouse/staff/performance')
    .then((r) => r.data as { staff_performance: StaffPerformanceRow[] })

export const getMyPerformance = () =>
  apiClient.get('/warehouse/staff/performance/me').then((r) => r.data as StaffPerformanceRow)

// ---- Warehouse locations ----

export const listZones = () =>
  apiClient.get('/warehouse/zones').then((r) => r.data as { zones: WarehouseZone[] })

export const createZone = (name: string) =>
  apiClient.post('/warehouse/zones', { name }).then((r) => r.data as WarehouseZone)

export const listRacks = (zoneId: number) =>
  apiClient.get(`/warehouse/zones/${zoneId}/racks`).then((r) => r.data as { racks: WarehouseRack[] })

export const createRack = (zoneId: number, name: string) =>
  apiClient.post(`/warehouse/zones/${zoneId}/racks`, { name }).then((r) => r.data as WarehouseRack)

export const listBins = (rackId: number) =>
  apiClient.get(`/warehouse/racks/${rackId}/bins`).then((r) => r.data as { bins: WarehouseBin[] })

export const createBin = (rackId: number, name: string) =>
  apiClient.post(`/warehouse/racks/${rackId}/bins`, { name }).then((r) => r.data as WarehouseBin)

export const getProductInventory = (productId: number) =>
  apiClient.get(`/warehouse/inventory/${productId}`).then((r) => r.data as Inventory)

export const assignProductBin = (productId: number, binId: number | null) =>
  apiClient
    .put(`/warehouse/inventory/${productId}/bin`, { bin_id: binId })
    .then((r) => r.data as Inventory)

// ---- Stock adjustment / movement ----

export const adjustStock = (
  productId: number,
  data: { new_quantity: number; reason: string; notes?: string }
) => apiClient.post(`/warehouse/inventory/${productId}/adjust`, data).then((r) => r.data as Inventory)

export const listStockMovements = (params: { product_id?: number; movement_type?: string; page?: number; limit?: number }) =>
  apiClient.get('/warehouse/stock-movements', { params }).then((r) => r.data as StockMovementsResponse)

// ---- Warehouse location deletion ----

export const deleteZone = (zoneId: number) =>
  apiClient.delete(`/warehouse/zones/${zoneId}`).then((r) => r.data)

export const deleteRack = (rackId: number) =>
  apiClient.delete(`/warehouse/racks/${rackId}`).then((r) => r.data)

export const deleteBin = (binId: number) =>
  apiClient.delete(`/warehouse/bins/${binId}`).then((r) => r.data)

// ---- Barcode scan ----

export const scanPickItem = (itemId: number, barcode: string) =>
  apiClient.put(`/warehouse/picking/items/${itemId}/scan`, { barcode }).then((r) => r.data as ScanResult)
