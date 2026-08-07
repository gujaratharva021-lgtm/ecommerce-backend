export interface WarehouseStaff {
  id: number
  name: string
  phone: string
  warehouse_id: number
  is_active?: boolean
  warehouse?: Warehouse
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

export interface Product {
  id: number
  name: string
  price: number
}

export interface StockTransfer {
  id: number
  product_id: number
  product?: Product
  from_warehouse_id: number
  from_warehouse?: Warehouse
  to_warehouse_id: number
  to_warehouse?: Warehouse
  quantity: number
  status: 'pending' | 'in_transit' | 'received' | 'rejected'
  requested_by: number
  approved_by?: number | null
  created_at?: string
}
