import apiClient from './client'
import type { Product, ProductCreateRequest, CreateCouponRequest, CreateOfferRequest, DeliveryPartner, Warehouse, WarehouseStaff } from '../types/admin'

export const IMAGE_ORIGIN = (apiClient.defaults.baseURL ?? '').replace(/\/api\/v1\/?$/, '')

export const uploadImage = (file: File) => {
  const formData = new FormData()
  formData.append('image', file)
  return apiClient
    .post('/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    .then((r) => r.data as { image_url: string })
}

// ---- Products ----
export const listProducts = (params?: Record<string, any>) =>
  apiClient.get('/products', { params }).then((r) => r.data)

export const createProduct = (data: ProductCreateRequest) =>
  apiClient.post('/admin/products', data).then((r) => r.data)

export const updateProduct = (id: number, data: Partial<Product>) =>
  apiClient.put(`/admin/products/${id}`, data).then((r) => r.data)

export const deleteProduct = (id: number) =>
  apiClient.delete(`/admin/products/${id}`).then((r) => r.data)

export const updateInventory = (id: number, stock: number, warehouseId: number) =>
  apiClient.put(`/admin/products/${id}/inventory`, { stock, warehouse_id: warehouseId }).then((r) => r.data)

// ---- Categories ----
export const listCategories = () =>
  apiClient.get('/categories').then((r) => r.data)

export const createCategory = (name: string) =>
  apiClient.post('/admin/categories', { name }).then((r) => r.data)

export const updateCategory = (id: number, name: string) =>
  apiClient.put(`/admin/categories/${id}`, { name }).then((r) => r.data)

export const deleteCategory = (id: number) =>
  apiClient.delete(`/admin/categories/${id}`).then((r) => r.data)

// ---- Orders ----
export const listOrders = (params?: Record<string, any>) =>
  apiClient.get('/admin/orders', { params }).then((r) => r.data)

export const updateOrderStatus = (id: number, status: string) =>
  apiClient.put(`/admin/orders/${id}/status`, { status }).then((r) => r.data)

export const assignDeliveryPartner = (orderId: number, deliveryPartnerId: number) =>
  apiClient.put(`/admin/orders/${orderId}/assign-delivery`, { delivery_partner_id: deliveryPartnerId }).then((r) => r.data)

// ---- Coupons ----
export const listCoupons = () =>
  apiClient.get('/admin/coupons').then((r) => r.data)

export const createCoupon = (data: CreateCouponRequest) =>
  apiClient.post('/admin/coupons', data).then((r) => r.data)

export const updateCouponStatus = (id: number, isActive: boolean) =>
  apiClient.put(`/admin/coupons/${id}/status`, { is_active: isActive }).then((r) => r.data)

// ---- Delivery Partners ----
export const listDeliveryPartners = () =>
  apiClient.get('/admin/delivery-partners').then((r) => r.data)

export const createDeliveryPartner = (data: Partial<DeliveryPartner>) =>
  apiClient.post('/admin/delivery-partners', data).then((r) => r.data)

export const updateDeliveryPartner = (id: number, data: Partial<DeliveryPartner>) =>
  apiClient.put(`/admin/delivery-partners/${id}`, data).then((r) => r.data)

export const deleteDeliveryPartner = (id: number) =>
  apiClient.delete(`/admin/delivery-partners/${id}`).then((r) => r.data)

// ---- Warehouses ----
export const listWarehouses = () =>
  apiClient.get('/admin/warehouses').then((r) => r.data)

export const getWarehouse = (id: number) =>
  apiClient.get(`/admin/warehouses/${id}`).then((r) => r.data)

export const createWarehouse = (data: Partial<Warehouse>) =>
  apiClient.post('/admin/warehouses', data).then((r) => r.data)

export const updateWarehouse = (id: number, data: Partial<Warehouse>) =>
  apiClient.put(`/admin/warehouses/${id}`, data).then((r) => r.data)

export const deleteWarehouse = (id: number) =>
  apiClient.delete(`/admin/warehouses/${id}`).then((r) => r.data)

export const setWarehouseServiceArea = (id: number, geojson: string) =>
  apiClient.put(`/admin/warehouses/${id}/service-area`, { geojson }).then((r) => r.data)

// ---- Warehouse Staff ----
export const listWarehouseStaff = () =>
  apiClient.get('/admin/warehouse-staff').then((r) => r.data)

export const createWarehouseStaff = (data: Partial<WarehouseStaff>) =>
  apiClient.post('/admin/warehouse-staff', data).then((r) => r.data)

export const updateWarehouseStaff = (id: number, data: Partial<WarehouseStaff>) =>
  apiClient.put(`/admin/warehouse-staff/${id}`, data).then((r) => r.data)

export const deleteWarehouseStaff = (id: number) =>
  apiClient.delete(`/admin/warehouse-staff/${id}`).then((r) => r.data)

// ---- Stock Transfers ----
export const listStockTransfers = () =>
  apiClient.get('/admin/stock-transfers').then((r) => r.data)

export const approveStockTransfer = (id: number) =>
  apiClient.put(`/admin/stock-transfers/${id}/approve`).then((r) => r.data)

export const rejectStockTransfer = (id: number) =>
  apiClient.put(`/admin/stock-transfers/${id}/reject`).then((r) => r.data)

// ---- Returns ----
export const listReturns = () =>
  apiClient.get('/admin/returns').then((r) => r.data)

export const approveReturn = (id: number) =>
  apiClient.put(`/admin/returns/${id}/approve`).then((r) => r.data)

export const rejectReturn = (id: number) =>
  apiClient.put(`/admin/returns/${id}/reject`).then((r) => r.data)

// ---- Analytics ----
export const getAnalyticsSummary = () =>
  apiClient.get('/admin/analytics/summary').then((r) => r.data)

export const getProductPerformance = () =>
  apiClient.get('/admin/analytics/products').then((r) => r.data)

// ---- Wallet ----
export const creditWallet = (userId: number, amount: number, note?: string) =>
  apiClient.post(`/admin/wallet/credit/${userId}`, { amount, note }).then((r) => r.data)



export const cancelStockTransfer = (id: number) =>
  apiClient.put(`/admin/stock-transfers/${id}/cancel`).then((r) => r.data)

// ---- Customers ----
export const listCustomers = (params?: Record<string, any>) =>
  apiClient.get('/admin/customers', { params }).then((r) => r.data)

export const getCustomer = (id: number) =>
  apiClient.get(`/admin/customers/${id}`).then((r) => r.data)

export const blockCustomer = (id: number) =>
  apiClient.put(`/admin/customers/${id}/block`).then((r) => r.data)

export const unblockCustomer = (id: number) =>
  apiClient.put(`/admin/customers/${id}/unblock`).then((r) => r.data)

// ---- Inventory Overview ----
export const getInventoryOverview = (params?: Record<string, any>) =>
  apiClient.get('/admin/inventory', { params }).then((r) => r.data)

// ---- Staff & Roles ----
export const listAdminStaff = () =>
  apiClient.get('/admin/staff').then((r) => r.data)

export const updateStaffRole = (id: number, adminRole: string) =>
  apiClient.put(`/admin/staff/${id}/role`, { admin_role: adminRole }).then((r) => r.data)

// ---- Settings ----
export const getSettings = () =>
  apiClient.get('/admin/settings').then((r) => r.data)

export const updateSettings = (settings: Record<string, string>) =>
  apiClient.put('/admin/settings', { settings }).then((r) => r.data)

export const deleteCoupon = (id: number) =>
  apiClient.delete(`/admin/coupons/${id}`).then((r) => r.data)

// ---- Audit Logs ----
export const getAuditLogs = (params?: Record<string, any>) =>
  apiClient.get('/admin/audit-logs', { params }).then((r) => r.data)

// ---- Notifications ----
export const broadcastNotification = (title: string, body: string) =>
  apiClient.post('/admin/notifications/broadcast', { title, body }).then((r) => r.data)

// ---- Offers ----
export const listOffers = () =>
  apiClient.get('/admin/offers').then((r) => r.data)

export const createOffer = (data: CreateOfferRequest) =>
  apiClient.post('/admin/offers', data).then((r) => r.data)

export const updateOfferStatus = (id: number, isActive: boolean) =>
  apiClient.put(`/admin/offers/${id}/status`, { is_active: isActive }).then((r) => r.data)

export const deleteOffer = (id: number) =>
  apiClient.delete(`/admin/offers/${id}`).then((r) => r.data)
