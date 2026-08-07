export interface Product {
  id: number
  name: string
  description?: string
  price: number
  category_id: number
  image_url?: string
  created_at?: string
  inventories?: { id: number; warehouse_id: number; stock: number; in_stock: boolean }[]
}

export interface ProductCreateRequest {
  name: string
  description?: string
  price: number
  category_id: number
  image_url?: string
  stock?: number
}

export interface Category {
  id: number
  name: string
}

export interface Order {
  id: number
  user_id: number
  status: string
  total_amount: number
  payment_method?: string
  created_at: string
}

export interface Coupon {
  id: number
  code: string
  discount_type: "flat" | "percentage"
  discount_value: number
  min_order_amount: number
  max_discount_amount?: number | null
  usage_limit: number
  used_count: number
  expiry_date: string
  is_active: boolean
}

export interface CreateCouponRequest {
  code: string
  discount_type: "flat" | "percentage"
  discount_value: number
  min_order_amount?: number
  max_discount_amount?: number | null
  usage_limit?: number
  expiry_date: string
}

export interface DeliveryPartner {
  id: number
  name: string
  phone: string
  vehicle_number?: string
  is_active?: boolean
}

export interface Warehouse {
  id: number
  name: string
  city: string
  address?: string
  lat: number
  lng: number
  service_radius_km?: number
  is_active?: boolean
}

export interface WarehouseStaff {
  id: number
  name: string
  phone: string
  warehouse_id: number
  status?: string
}

export interface StockTransfer {
  id: number
  product_id: number
  from_warehouse_id: number
  to_warehouse_id: number
  quantity: number
  status: string
  created_at?: string
}

export interface ReturnRequest {
  id: number
  order_id: number
  reason?: string
  status: string
  created_at?: string
}

export interface AnalyticsSummary {
  total_users: number
  total_orders: number
  total_sales: number
  pending_orders: number
  delivered_orders: number
  cancelled_orders: number
}

export interface ProductPerformance {
  product_id: number
  product_name: string
  units_sold: number
  total_revenue: number
}

export interface PaginatedResponse<_T> {
  page: number
  limit: number
  total: number
  total_pages: number
  [key: string]: any // covers "products", "orders", etc. - key name varies by endpoint
}

export interface Wallet {
  id: number
  user_id: number
  balance: number
  created_at: string
  updated_at: string
}
