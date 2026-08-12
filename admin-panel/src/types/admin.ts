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

export interface OrderItem {
  id: number
  order_id: number
  product_id: number
  product?: Product
  quantity: number
  price: number
  created_at?: string
}

export interface Order {
  id: number
  user_id: number
  status: string
  total_amount: number
  payment_method?: string
  created_at: string
  items?: OrderItem[]
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
  service_area?: string
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

// ---- Customers ----
export interface CustomerSummary {
  id: number
  name: string
  phone: string
  is_blocked: boolean
  created_at: string
  total_orders: number
  total_spent: number
  last_order_at?: string | null
}

export interface CustomerDetail extends CustomerSummary {
  orders: Order[]
  addresses: {
    id: number
    label: string
    full_name: string
    phone: string
    line1: string
    line2?: string
    city: string
    state: string
    pincode: string
    is_default: boolean
  }[]
  wallet?: { id: number; balance: number } | null
  wallet_transactions?: { id: number; type: string; amount: number; created_at: string }[]
}

// ---- Inventory Overview ----
export interface InventoryRow {
  product_id: number
  product_name: string
  category_name: string
  warehouse_id: number
  warehouse_name: string
  stock: number
  reserved: number
  available: number
  in_stock: boolean
}

export interface InventoryOverviewResponse {
  total_skus: number
  total_available_stock: number
  total_reserved_stock: number
  low_stock_count: number
  out_of_stock_count: number
  damaged_stock: number
  expired_stock: number
  rows: InventoryRow[]
  page: number
  limit: number
  total: number
  total_pages: number
}

// ---- Staff & Roles ----
export interface StaffMember {
  id: number
  name: string
  phone: string
  admin_role?: string
  created_at: string
}

export const ADMIN_ROLES = [
  { value: 'super_admin', label: 'Super Admin' },
  { value: 'admin', label: 'Admin' },
  { value: 'ops_manager', label: 'Operations Manager' },
  { value: 'warehouse_manager', label: 'Warehouse Manager' },
  { value: 'support_agent', label: 'Support Agent' },
  { value: 'finance_manager', label: 'Finance Manager' },
]

// ---- Settings ----
export interface SettingsMap {
  [key: string]: string
}

export const SETTINGS_LABELS: Record<string, string> = {
  free_delivery_threshold: 'Free Delivery Threshold (â‚¹)',
  flat_delivery_charge: 'Flat Delivery Charge (â‚¹)',
  min_order_amount: 'Minimum Order Amount (â‚¹)',
  cancellation_window_minutes: 'Cancellation Window (minutes)',
  company_name: 'Company Name',
  support_phone: 'Support Phone',
  support_email: 'Support Email',
  gst_percentage: 'GST Percentage (%)',
}


// ---- Audit Logs ----
export interface AuditLog {
  id: number
  admin_id: number
  admin_phone: string
  action: string
  entity_type: string
  entity_id: string
  details: string
  created_at: string
}

export interface AuditLogsResponse {
  logs: AuditLog[]
  page: number
  limit: number
  total: number
  total_pages: number
}

// ---- Offers ----
export interface Offer {
  id: number
  title: string
  description: string
  image_url: string
  discount_text: string
  category_id?: number | null
  start_date: string
  end_date: string
  is_active: boolean
  created_at: string
}

export interface CreateOfferRequest {
  title: string
  description?: string
  image_url?: string
  discount_text?: string
  category_id?: number | null
  start_date: string
  end_date: string
}

// ---- Banners ----
export interface Banner {
  id: number
  image_url: string
  title: string
  link_type: 'product' | 'category' | 'url' | 'none'
  link_value: string
  display_order: number
  is_active: boolean
  created_at: string
}

export interface CreateBannerRequest {
  image_url: string
  title?: string
  link_type?: string
  link_value?: string
  display_order?: number
}
